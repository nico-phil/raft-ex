package main

import "flag"

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	distributedFSM, err := NewDistributedFSM()
	if err != nil {
		panic(err)
	}
	server := NewServer(*port, distributedFSM)

	err = server.Start()
	if err != nil {
		panic(err)
	}
}
