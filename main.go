package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	nodeID := flag.String("node-id", "", "unique node ID")
	raftAddr := flag.String("raft-addr", "", "Raft TCP address")
	httpAddr := flag.String("http-addr", "", "HTTP API address")
	bootstrap := flag.Bool("bootstrap", false, "should boostrap the cluster")
	joinAddr := flag.String("join-addr", "", "address to join the cluster")
	dataDir := flag.String("data-dir", "", "data dir to store raft logs")
	flag.Parse()

	fmt.Println("JOIN_ADRESS:", *joinAddr)

	err := os.MkdirAll(*dataDir, 0o755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatalf("failed datad-dir: %v", err)
		}

	}
	config := Config{
		NodeID:    *nodeID,
		Bootstrap: *bootstrap,
		JoinAddr:  *joinAddr,
		RaftAddr:  *raftAddr,
		DataDir:   *dataDir,
	}

	distributedFSM, err := NewDistributedFSM(config)
	if err != nil {
		panic(err)
	}
	fmt.Println("BEFORE")
	server := NewServer(
		distributedFSM,
		*nodeID,
		*raftAddr,
		*httpAddr,
		*joinAddr,
	)

	fmt.Println("AFTER")

	go func() {
		err = server.Start()
		if err != nil {
			panic(err)
		}
	}()

	if *joinAddr != "" {
		if err := server.JoinCluster(); err != nil {
			log.Fatalf("main:failed to join cluster: %v", err)
		}
	}

	select {}

}
