# Raft-ex

## What is Raft-ex?

Raft-ex is a minimal implementation of the Raft consensus algorithm using the excellent [HashiCorp Raft library](https://github.com/hashicorp/raft.git).

Raft is a distributed consensus protocol that allows a cluster of nodes to agree on the state of a replicated state machine, even in the presence of node failures or temporary network partitions. It is one of the fundamental building blocks of fault-tolerant distributed systems.

The goal of this project is not to be production-ready. Instead, it focuses on the smallest set of concepts needed to build an intuition for how Raft works and how it can be integrated into your own applications.

The replicated state machine is intentionally simple: a single integer value stored in memory. This keeps the implementation small while demonstrating the complete Raft workflow:
- leader election
- log replication
- state machine application
- node membership
- automatic recovery after failures

## Architecture
Each node exposes two network endpoints:

Raft address (raft-addr) — used exclusively for Raft communication between nodes over TCP.
HTTP address (http-addr) — used by clients to interact with the cluster.

Client requests are sent to the leader through the HTTP API. The leader appends each request to the Raft log, replicates it to the followers, and only applies it to the state machine once the entry has been committed by a majority of the cluster.

                Client
                   │
             HTTP /set
                   │
             ┌───────────┐
             │  Leader   │
             └───────────┘
                   │
         AppendEntries (TCP)
         ┌─────────┴─────────┐
         ▼                   ▼
    Follower            Follower
         │                   │
         └──── Apply Log ────┘


## Building Raft-ex
Building Raft-ex requires Go 1.20 or later
```bash
git clone https://github.com/nico-phil/raft-ex.git
cd raft-ex
go install
```

## Runnning a cluster
Each node requires its own configuration.
`node-id`: Unique node identifier 
`raft-addr`: TCP address used by Raft
`http-addr`: HTTP API address
`data-dir`: Directory for Raft logs and snapshots
`bootstrap`: Bootstrap a brand-new cluster (only one node)
`join-addr`: HTTP address of an existing cluster member

**Important**
- Only one node should bootstrap the cluster.
- Every other node joins the existing cluster using --join-addr.

### Start the first node like so:
```bash
go run *.go \
--node-id=node-0 \
--raft-addr=127.0.0.1:7000 \
--http-addr=127.0.0.1:8000 \
--data-dir=node-0 \
--bootstrap=true
```
Once started, this node becomes the initial cluster leader.

### Add 2 more nodes
With three voting nodes, the cluster can tolerate the failure of one node while continuing to process requests.

`node-1`:
```bash
go run *.go \
--node-id=node-1 \
--raft-addr=127.0.0.1:7001 \
--http-addr=127.0.0.1:8001 \
--join-addr=127.0.0.1:8000 \
--data-dir=node-1 \
```

`node-2`:
```bash
go run *.go \
--node-id=node-2 \
--raft-addr=127.0.0.1:7002 \
--http-addr=127.0.0.1:8002 \
--join-addr=127.0.0.1:8000 \
--data-dir=node-2 \
```

## Cluster Membership
You can inspect the current cluster membership by querying any node:
```bash
curl -i -X GET http://127.0.0.1:8000/getservers
```
Example output:
```bash
- node_id:node-0, raft_addr:127.0.0.1:7000, leader:true 
- node_id:node-1, raft_addr:127.0.0.1:7001, leader:false 
- node_id:node-2, raft_addr:127.0.0.1:7002, leader:false
```

## Reading and Updating the State Machine
The state machine consists of a single replicated integer value. You set a value by sending an http request the leader node like so: 

```bash
curl \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"value":20}' \
  http://127.0.0.1:8000/set
```

The leader will:
1. Append the command to its Raft log.
2. Replicate the log entry to the followers.
3. Wait until a majority acknowledges it.
4. Apply the command to the state machine on every node.


### Read the Value
Query any node
```bash
curl -i -X GET http://127.0.0.1:8000/get 
curl -i -X GET http://127.0.0.1:8001/get 
curl -i -X GET http://127.0.0.1:8002/get 
```
All nodes should return the same value.

## Fault tolerance

If the leader node fails(simlutate failure with`Ctrl+C`). Within a short period of time, one of the remaining nodes will automatically be elected as the new leader.Client requests can then continue to be sent to the newly elected leader. 
Later, When the first node is restarted it will rejoin the cluster and gets updated with any new state while it was down

## What This Project Demonstrates

Although intentionally minimal, this project includes the essential pieces of a Raft-based replicated state machine:

- Leader election
- Log replication
- State machine application
- Dynamic node joins
- Automatic leader failover
- Recovery of restarted nodes
- Strong consistency through majority consensus

By replacing the simple in-memory integer with your own state machine, you can use the same architecture to build replicated key-value stores, metadata services, distributed schedulers, and many other fault-tolerant distributed systems.

## References 
-  Raft paper, [In Search of an Understandable Consensus Algorithm (Extended Version) by Diego Ongaro and John Ousterhout](https://raft.github.io/raft.pdf)
- GopherCon2023 talk "Build Your Own Distributed System Using Go" ([video](https://www.youtube.com/watch?v=8XbxQ1Epi5w), [slides](https://storage.googleapis.com/bucket-vallified/BuildYourOwnDistributedSystemUsingGoGopherCon2023.pdf)).

- hraftd, a reference example use of the Hashicorp Raft implementation [github](https://github.com/otoolep/hraftd).

- Practical Distributed Consensus using HashiCorp/raft [video](https://www.youtube.com/watch?v=EGRmmxVFOfE&t=958s).