package main

import "flag"

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	nodeID := flag.String("nodeID", "0", "id of the node")
	bootstrap := flag.Bool("bootstrap", false, "should boostrap the cluster")
	joinAddr := flag.String("joinAddr", "", "address to join the cluster")
	flag.Parse()

	config := Config{
		NodeID:    *nodeID,
		Bootstrap: *bootstrap,
		JoinAddr:  *joinAddr,
	}

	distributedFSM, err := NewDistributedFSM(config)
	if err != nil {
		panic(err)
	}
	server := NewServer(*port, distributedFSM)

	err = server.Start()
	if err != nil {
		panic(err)
	}
}
