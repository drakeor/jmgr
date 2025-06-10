# About JMGR

-Work in progress- 

JMGR is an edge-focused fault-tolerant telemetry tool for real-time monitoring and alerting across a cluster of servers. Each monitored node runs a lightweight agent that collects key metrics and participates in decentralized data synchronization with its peers. The system uses a custom anti-entropy protocol for robust replication and comes with an optional front-end for viewing the state of the cluster.

This is a work in progress.

# Prerequisites
```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go get github.com/dgraph-io/badger/v4
go get github.com/shirou/gopsutil/v4
go get github.com/shirou/gopsutil/v4/internal/common@v4.25.4

go get github.com/tklauser/go-sysconf
```

# Quick Start

Install the prerequistites and build the test cluster using the following command:

```bash
make local
```

# Side notes


To get protobuf working:

```bash
export GO_PATH=~/go
export PATH=$PATH:/$GO_PATH/bin
```

To run specific test: go test -v -race ./internal/antientropy