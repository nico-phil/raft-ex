package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	fsm *FSM
}

func NewServer() *Server {
	return &Server{
		fsm: NewFSM(),
	}
}

func (s *Server) Start() error {
	http.HandleFunc("POST /add", s.handleAdd)
	http.HandleFunc("GET /get", s.handleGet)

	log.Println("Server is starting on port 8080...")
	return http.ListenAndServe(":8080", nil)
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
