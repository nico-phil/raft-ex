package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	fsm  *DistributedFSM //*FSM
	port int
}

func NewServer(port int, d *DistributedFSM) *Server {
	return &Server{
		fsm:  d,
		port: port,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("POST /add", s.handleAdd)
	http.HandleFunc("GET /get", s.handleGet)

	log.Printf("Server is starting on port %d...", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
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
