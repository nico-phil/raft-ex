package main

import (
	"encoding/json"
	"fmt"
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

type event struct {
	Type  string
	Value int
}

func (f *FSM) Apply(l *raft.Log) interface{} {
	var e event
	err := json.Unmarshal(l.Data, &e)
	if err != nil {
		return err
	}

	switch e.Type {
	case "Add":
		f.Add(e.Value)
	default:
		fmt.Println("Unknow event:", e.Type)
	}

	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.m.Lock()
	defer f.m.Unlock()
	s := &fsmSnapshot{StateValue: f.stateValue}
	return s, nil
}

func (f *FSM) Restore(snapshot io.ReadCloser) error {
	var sn fsmSnapshot
	err := json.NewDecoder(snapshot).Decode(&sn)
	if err != nil {
		return err
	}

	f.stateValue = sn.StateValue
	return nil
}

type fsmSnapshot struct {
	StateValue int
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (s *fsmSnapshot) Release() {}
