package raftnode

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"fmt"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

var Raft *raft.Raft
var fsm *RaftFSM

func GetFSM() *RaftFSM {
	return fsm
}

func IsLeader() bool {
	return Raft.State() == raft.Leader
}

func GetLeaderAddress() string {
	return string(Raft.Leader())
}

func ApplyCommand(commandType string, payload any) error {
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(payload); err != nil {
		return err
	}

	cmd := Command{
		Type: commandType,
		Data: data.Bytes(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(cmd); err != nil {
		return err
	}

	future := Raft.Apply(buf.Bytes(), 5*time.Second)
	return future.Error()
}

func StartRaftNode(nodeID, raftAddr, peerStr string, bootstrap bool) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	config.HeartbeatTimeout = 500 * time.Millisecond
	config.ElectionTimeout = 1000 * time.Millisecond
	config.LeaderLeaseTimeout = 500 * time.Millisecond
	config.CommitTimeout = 50 * time.Millisecond

	raftDir := filepath.Join("raft-data", nodeID)
	os.MkdirAll(raftDir, 0700)

	logStore, err := raft.NewLogCache(512, raft.NewInmemStore())
	if err != nil {
		log.Fatalf("failed to create log store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft.db"))
	if err != nil {
		log.Fatalf("failed to create stable store: %v", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		log.Fatalf("failed to create snapshot store: %v", err)
	}

	transport, err := raft.NewTCPTransport(raftAddr, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatalf("failed to create transport: %v", err)
	}

	fsm = NewFSM()
	raftNode, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		log.Fatalf("failed to create raft node: %v", err)
	}
	Raft = raftNode

	if bootstrap {
		log.Println("Bootstrapping cluster...")
		peerList := strings.Split(peerStr, ",")
		var servers []raft.Server
		servers = append(servers, raft.Server{ID: raft.ServerID(nodeID), Address: raft.ServerAddress(raftAddr)})
		for i, peer := range peerList {
			peerID := fmt.Sprintf("node%d", i+2)
			servers = append(servers, raft.Server{
				ID:      raft.ServerID(peerID),
				Address: raft.ServerAddress(peer),
			})
		}
		future := Raft.BootstrapCluster(raft.Configuration{Servers: servers})
		if err := future.Error(); err != nil {
			log.Fatalf("failed to bootstrap cluster: %v", err)
		}
	}
}
