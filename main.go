package main

import (
	"flag"
	"log"
	"net/http"
	"Raft3D/internal/raftnode"
	"github.com/gorilla/mux"
	"Raft3D/internal/api"
)

func main() {
	id := flag.String("id", "", "Node ID")
	httpPort := flag.String("http", "127.0.0.1:8000", "HTTP bind address")
	raftPort := flag.String("raft", "", "Raft bind address (e.g., 127.0.0.1:9000)")
	peers := flag.String("peers", "", "Comma-separated list of peer raft addresses")
	bootstrap := flag.Bool("bootstrap", false, "Bootstrap the cluster")

	flag.Parse()

	if *id == "" || *raftPort == "" || *peers == "" {
		log.Fatal("All flags --id, --raft, and --peers are required")
	}

	raftnode.StartRaftNode(*id, *raftPort, *peers, *bootstrap)

	router := mux.NewRouter()
	api.RegisterRoutes(router)
	// Start HTTP server (you can add handler wiring here)
	log.Printf("Server running on port %s\n", *httpPort)
	log.Fatal(http.ListenAndServe(*httpPort, router))
}
