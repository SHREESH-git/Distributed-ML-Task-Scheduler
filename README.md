# Distributed ML Task Scheduler

A lightweight, distributed task scheduling system built in Go. Designed to delegate heavy data processing and machine learning inference workloads across multiple remote worker nodes to eliminate single-node bottlenecks.

## Current Implementation (MVP)
This project is currently in active development. The foundational networking architecture is established:
* **Master-Worker Topology:** Core node architecture implemented via the `gorilla/websocket` package.
* **Connection Handshake:** Master node successfully hosts the WebSocket server, while Worker nodes can dynamically connect, register their readiness status, and stand by for task delegation.

## Development Roadmap 
Upcoming infrastructure upgrades to reach production readiness:
- [ ] **Dynamic Load Balancing:** Implement a round-robin algorithm on the Master Node to evenly distribute inference workloads across all active workers.
- [ ] **Task Queues:** Add an in-memory queue to hold pending tasks if all workers are currently busy.
- [ ] **Fault Tolerance (Heartbeats):** Implement ping/pong heartbeat checks to detect dead workers and automatically re-queue their assigned tasks.

## Quick Start (Local Testing)
1. Start the Master Node: `go run master.go`
2. Start a Worker Node: `go run worker.go`