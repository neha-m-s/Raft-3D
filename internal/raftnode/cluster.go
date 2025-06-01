package raftnode

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

// NewRaft initializes a Raft node
func NewRaft(nodeID, raftDir string) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	store, err := raftboltdb.NewBoltStore(fmt.Sprintf("%s/log.db", raftDir))
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return nil, err
	}

	trans, err := raft.NewTCPTransport(":12000", nil, 3, raftTimeout, os.Stderr)
	if err != nil {
		return nil, err
	}

	raftNode, err := raft.NewRaft(config, NewFSM(), store, store, snapshotStore, trans)
	if err != nil {
		return nil, err
	}

	log.Println("Raft node started successfully:", nodeID)
	return raftNode, nil
}
