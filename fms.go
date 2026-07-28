package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/hashicorp/raft"
)

// FSM represents the finite-date machine(the database)
// in this example, i use just an integer to represent the db for better mental model.
// the finite state machine can anything that holds state: file system, postgres, in memory data...
type FSM struct {
	m          sync.Mutex
	stateValue int
}

// NewFSM returns a new finite state machine
func NewFSM() *FSM {
	return &FSM{}
}

// Add modifies the value of the state
func (f *FSM) Add(value int) {
	f.m.Lock()
	defer f.m.Unlock()
	f.stateValue = value
}

// Get return the state value
func (f *FSM) Get() int {
	f.m.Lock()
	defer f.m.Unlock()
	return f.stateValue
}

type event struct {
	Type  string
	Value int
}

// Apply: this method is required by raft,
// it's this method that get executed to modify the state once a log entry is committed by a majority of the cluster.
func (f *FSM) Apply(l *raft.Log) interface{} {
	var e event
	err := json.Unmarshal(l.Data, &e)
	if err != nil {
		return err
	}

	switch e.Type {
	case "Add":
		f.Add(e.Value)
		log.Printf(
			"FSM applied: node log_index=%d term=%d value=%d",
			l.Index,
			l.Term,
			e.Value,
		)
	default:
		return fmt.Errorf("unknown event type: %q", e.Type)
	}

	return nil
}

// Snapshot is required by raft. It return a snapshot fo the state
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.m.Lock()
	defer f.m.Unlock()
	s := &fsmSnapshot{StateValue: f.stateValue}
	return s, nil
}

// Restore is required by raft. it restores the state from the snapshot
func (f *FSM) Restore(snapshot io.ReadCloser) error {
	var sn fsmSnapshot
	err := json.NewDecoder(snapshot).Decode(&sn)
	if err != nil {
		return err
	}

	f.stateValue = sn.StateValue
	return nil
}

// fsmSnapshot represent a snapshot of the fsm.
type fsmSnapshot struct {
	StateValue int
}

// Persist is required by raft
func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

// Release is required by raft
func (s *fsmSnapshot) Release() {}
