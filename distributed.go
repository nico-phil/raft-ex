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

type Config struct {
	NodeID    string
	Bootstrap bool
	JoinAddr  string
}

func NewDistributedFSM(config Config) (*DistributedFSM, error) {
	dFMS := DistributedFSM{
		fsm: NewFSM(),
	}

	var err error
	err = dFMS.setupRaft(config)
	if err != nil {
		return nil, err
	}

	return &dFMS, nil
}

func (d *DistributedFSM) setupRaft(config Config) error {
	raftconfig := raft.DefaultConfig()
	raftconfig.LocalID = raft.ServerID(config.NodeID)

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

	d.raft, err = raft.NewRaft(raftconfig, d.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return err
	}

	if config.Bootstrap {

		d.raft.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: raft.ServerID(config.NodeID), Address: transport.LocalAddr()},
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

func (d *DistributedFSM) Join(id, addr string) error {

	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)

	indexFuture := d.raft.AddVoter(serverID, serverAddr, 0, 0)
	if err := indexFuture.Error(); err != nil {
		return err
	}

	return nil
}
