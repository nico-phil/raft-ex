package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/raft"
)

// Node represent a node in the cluster
type Node struct {
	id       string
	raftAddr string
	httpAddr string
	joinAddr string
	fsm      *DistributedFSM
}

// NewNode creates and returns new node
func NewNode(
	d *DistributedFSM,
	id string,
	raftAddr string,
	httpAddr string,
	joinAdrr string,

) *Node {
	return &Node{
		fsm:      d,
		id:       id,
		raftAddr: raftAddr,
		httpAddr: httpAddr,
		joinAddr: joinAdrr,
	}
}

// Start run the node server
func (s *Node) Start() error {
	http.HandleFunc("POST /add", s.handleSet)
	http.HandleFunc("GET /get", s.handleGet)
	http.HandleFunc("POST /join", s.handleJoin)
	http.HandleFunc("GET /getservers", s.handleGetServers)

	log.Printf("Server is starting on port %s...", s.httpAddr)

	return http.ListenAndServe(s.httpAddr, nil)
}

// handleSet handles set request, set the state
func (s *Node) handleSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value int `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle the add request here
	s.fsm.Set(req.Value)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Add request received"))
}

// handleGet return the state
func (s *Node) handleGet(w http.ResponseWriter, r *http.Request) {
	// Handle the get request here
	value := s.fsm.Get()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Current state value:" + strconv.Itoa(value) + "\n"))
}

// JoinRequest reprepsents a join request type
type JoinRequest struct {
	NodeID   string `json:"node_id"`
	RaftAddr string `json:"raft_addr"`
	HTTPAddr string `json:"http_addr"`
}

// handleJoin handles join request, allow node to join the cluster
func (s *Node) handleJoin(w http.ResponseWriter, r *http.Request) {
	var request JoinRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}

	log.Printf(
		"received join request: node_id=%q raft_addr=%q \n", request.NodeID, request.RaftAddr)

	if request.NodeID == "" || request.RaftAddr == "" {
		http.Error(
			w,
			"node_id and raft_addr are required",
			http.StatusBadRequest,
		)
		return
	}

	if s.fsm.raft.State() != raft.Leader {
		http.Error(w, "Node is not the leader", http.StatusBadRequest)
		return
	}

	if err := s.fsm.Join(request); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("node joined cluster"))

}

// JoinCluster: http request to join the cluster
func (s *Node) JoinCluster() error {

	client := http.Client{}
	joinReq := JoinRequest{
		NodeID:   s.id,
		RaftAddr: s.raftAddr,
		HTTPAddr: s.httpAddr,
	}

	body, err := json.Marshal(joinReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := "http://" + s.joinAddr + "/join"
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create NewRequest: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to Do request: %w", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to Join cluster: status=%d, body=%s", resp.StatusCode, string(responseBody))
	}
	return nil

}

// handleGetServers return all the nodes in the cluster
func (s *Node) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers := s.fsm.raft.GetConfiguration().Configuration().Servers
	fmt.Println("servers:")
	for _, server := range servers {
		fmt.Println("is_leader:", s.fsm.raft.Leader() == server.Address)
		fmt.Printf("\t- %v\n", server)
	}

	w.WriteHeader(http.StatusOK)
}
