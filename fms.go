package main

import "sync"

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
