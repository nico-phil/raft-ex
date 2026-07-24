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

type Server struct {
	id       string
	raftAddr string
	httpAddr string
	joinAddr string
	fsm      *DistributedFSM //*FSM

}

// type Node struct {
// 	id       raft.ServerID
// 	raftAddr raft.ServerAddress
// 	httpAddr string

// }

func NewServer(
	d *DistributedFSM,
	id string,
	raftAddr string,
	httpAddr string,
	joinAdrr string,

) *Server {
	return &Server{
		fsm:      d,
		id:       id,
		raftAddr: raftAddr,
		httpAddr: httpAddr,
		joinAddr: joinAdrr,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("POST /add", s.handleAdd)
	http.HandleFunc("GET /get", s.handleGet)
	http.HandleFunc("POST /join", s.handleJoin)

	log.Printf("Server is starting on port %s...", s.httpAddr)

	return http.ListenAndServe(s.httpAddr, nil)
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value int `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle the add request here
	s.fsm.Add(req.Value)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Add request received"))
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	// Handle the get request here
	value := s.fsm.Get()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Current state value:" + strconv.Itoa(value) + "\n"))
}

type JoinRequest struct {
	NodeID   string `json:"node_id"`
	RaftAddr string `json:"raft_addr"`
	HTTPAddr string `json:"http_addr"`
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	var request JoinRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}

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

func (s *Server) JoinCluster() error {

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
