package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	RaftAddr  string
	DataDir   string
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

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(config.DataDir, "raft-log.bolt"))
	if err != nil {
		return err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(config.DataDir, "raft-stable.bolt"))
	if err != nil {
		return err
	}

	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(config.DataDir, "raft-snapshot-store"),
		retain,
		os.Stderr,
	)

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.RaftAddr)
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

func (d *DistributedFSM) Add(value int) error {
	data := event{
		Type:  "Add",
		Value: value,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	applyFuture := d.raft.Apply(b, 10*time.Second)

	if err := applyFuture.Error(); err != nil {
		return err
	}

	return nil
}

func (d *DistributedFSM) Get() int {
	return d.fsm.Get()
}

func (d *DistributedFSM) Join(request JoinRequest) error {

	nodeID := raft.ServerID(request.NodeID)
	nodeAddr := raft.ServerAddress(request.RaftAddr)

	configureFuture := d.raft.GetConfiguration()
	if err := configureFuture.Error(); err != nil {
		return fmt.Errorf("cannot get raft configuration: %w", err)
	}

	for _, server := range configureFuture.Configuration().Servers {
		sameID := server.ID == nodeID
		sameAddress := server.Address == nodeAddr

		if sameID && sameAddress {
			log.Printf("node already belong to cluster: nodeID=%s, address=%s ", server.ID, server.Address)
			return nil
		}

		if (sameID && !sameAddress) || sameAddress && !sameID {
			log.Printf(
				"removing stale Raft server: id=%s address=%s", server.ID, server.Address,
			)

			removeFuture := d.raft.RemoveServer(server.ID, 0, 10*time.Second)
			if err := removeFuture.Error(); err != nil {
				return fmt.Errorf("failed to remove server: %w", err)
			}
		}

	}

	indexFuture := d.raft.AddVoter(nodeID, nodeAddr, 0, 10*time.Second)
	if err := indexFuture.Error(); err != nil {
		return fmt.Errorf("add voter: %w", err)
	}

	return nil
}
