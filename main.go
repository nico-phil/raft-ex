package main

import (
	"errors"
	"flag"
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

	err := os.MkdirAll(*dataDir, 0o755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatalf("failed datad-dir: %v", err)
		}
	}

	nodeConfig := NodeConfig{
		id:        *nodeID,
		raftAddr:  *raftAddr,
		httpAddr:  *httpAddr,
		joinAddr:  *joinAddr,
		dataDir:   *dataDir,
		bootstrap: *bootstrap,
	}

	distributedFSM, err := NewDistributedFSM(nodeConfig)
	if err != nil {
		panic(err)
	}

	server := NewNode(nodeConfig, distributedFSM)

	go func() {
		err = server.Start()
		if err != nil {
			panic(err)
		}
	}()

	if nodeConfig.joinAddr != "" {
		if err := server.JoinCluster(); err != nil {
			log.Fatalf("main:failed to join cluster: %v", err)
		}
	}

	select {}

}
