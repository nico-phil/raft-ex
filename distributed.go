package main

import (
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type DistributedFSM struct {
	fsm  *FSM
	raft *raft.Raft
}

func NewDistributedFSM() (*DistributedFSM, error) {
	dFMS := DistributedFSM{
		fsm: NewFSM(),
	}
	config := Config{
		nodeID:    "0",
		bootstrap: true,
	}
	var err error
	err = dFMS.setupRaft(config)
	if err != nil {
		return nil, err
	}

	return &dFMS, nil
}

type Config struct {
	nodeID    string
	bootstrap bool
}

func (d *DistributedFSM) setupRaft(c Config) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(c.nodeID)

	logStore, err := raftboltdb.NewBoltStore(filepath.Join("raft", "raft-log.bolt"))
	if err != nil {
		return err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join("raft", "raft-stable.bolt"))
	if err != nil {
		return err
	}

	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join("raft", "raft-snapshot-store"),
		retain,
		os.Stderr,
	)

	raftAddr := "localhost:8400"
	tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(tcpAddr.String(), tcpAddr, 10, time.Second*10, os.Stderr)

	d.raft, err = raft.NewRaft(config, d.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return err
	}

	if c.bootstrap {

		d.raft.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: raft.ServerID(c.nodeID), Address: transport.LocalAddr()},
			},
		})
	}

	return nil

}

func (d *DistributedFSM) Add(value int) {
	d.fsm.Add(value)
}

func (d *DistributedFSM) Get() int {
	return d.fsm.Get()
}
