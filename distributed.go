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

// DistributedFSM represents the distriubted fms
type DistributedFSM struct {
	fsm    *FSM
	raft   *raft.Raft
	config NodeConfig
}

// NewDistributedFSM returns a DistributedFSM
func NewDistributedFSM(config NodeConfig) (*DistributedFSM, error) {
	dFMS := DistributedFSM{
		fsm:    NewFSM(),
		config: config,
	}

	var err error
	err = dFMS.setupRaft()
	if err != nil {
		return nil, err
	}

	return &dFMS, nil
}

// setupRaft configures and creates a raft instance
func (d *DistributedFSM) setupRaft() error {
	raftconfig := raft.DefaultConfig()
	raftconfig.LocalID = raft.ServerID(d.config.id)

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(d.config.dataDir, "raft-log.bolt"))
	if err != nil {
		return err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(d.config.dataDir, "raft-stable.bolt"))
	if err != nil {
		return err
	}

	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(d.config.dataDir, "raft-snapshot-store"),
		retain,
		os.Stderr,
	)

	tcpAddr, err := net.ResolveTCPAddr("tcp", d.config.raftAddr)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(tcpAddr.String(), tcpAddr, 10, time.Second*10, os.Stderr)

	d.raft, err = raft.NewRaft(raftconfig, d.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return err
	}

	if d.config.bootstrap {

		d.raft.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: raft.ServerID(d.config.id), Address: transport.LocalAddr()},
			},
		})
	}

	return nil

}

// Set sets the state of fsm
func (d *DistributedFSM) Set(event event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	applyFuture := d.raft.Apply(b, 10*time.Second)

	if err := applyFuture.Error(); err != nil {
		return err
	}

	return nil
}

// Get returns the state value
func (d *DistributedFSM) Get() int {
	return d.fsm.Get()
}

// Join adds new node to the cluster. Only the leader can add new node
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
