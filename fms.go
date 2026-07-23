package main

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type FSM struct {
	m          sync.Mutex
	stateValue int
}

func NewFSM() *FSM {
	return &FSM{}
}

func (f *FSM) Add(value int) {
	f.m.Lock()
	defer f.m.Unlock()
	f.stateValue = value
}

func (f *FSM) Get() int {
	f.m.Lock()
	defer f.m.Unlock()
	return f.stateValue
}

type AddPayload struct {
	Value int
}

func (f *FSM) Apply(l *raft.Log) interface{} {
	var p AddPayload
	err := json.Unmarshal(l.Data, &p)
	if err != nil {
		return err
	}
	f.Add(p.Value)
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	s := &fsmSnapshot{}
	return s, nil
}

func (f *FSM) Restore(snapshot io.ReadCloser) error {
	return nil
}

type fsmSnapshot struct{}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (s *fsmSnapshot) Release() {}
