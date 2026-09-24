# Go Matchmaking Backend API

This repository contains a Go-based backend API that powers a real-time multiplayer matchmaking and social system inspired by FACEIT.

The backend is designed with clear internal layering (API, services, repositories), focusing on real-time communication, safe concurrency, and automated game server orchestration. It uses PostgreSQL, Redis, WebSockets, and gRPC.

---

## Features

### Authentication & Sessions
- Custom authentication system written in Go
- PostgreSQL-backed refresh token storage
- Secure login and registration flows (with CSRF protection)

### Matchmaking & Orchestration
- FIFO-based matchmaking logic
- Atomic match creation (Redis Lua scripts + Postgres `FOR UPDATE` locks)
- **Automated Game Server Provisioning**: Dynamically spins up AWS EC2 instances for matches
- **gRPC Sidecar Integration**: A custom Go sidecar runs alongside the Counter-Strike: Source (CS:S) server, streaming `console.log` events and heartbeats back to the orchestrator.
- Graceful server cleanup (`TerminateServer`) when matches conclude.

### Real-Time Communication
- WebSocket-based notifications hub
- Real-time party invites and matchmaking state updates sent directly to clients

### Social Features
- Friends and blocking system
- Party creation and management (role-based invites)
- Reporting system

---

## Tech Stack

- **Language:** Go 1.26
- **Database:** PostgreSQL 17
- **Cache / Queues:** Redis 7
- **Real-Time Web:** WebSockets
- **Internal Microservices:** gRPC (Backend ↔ Game Server Sidecar)
- **Infrastructure:** Docker, Docker Compose, AWS EC2, SteamCMD

---

## Architecture Overview

The backend operates as a core API server communicating with isolated game servers:

- **API Layer**: HTTP REST endpoints for authentication, matchmaking, and social features.
- **Service Layer**: Business logic (matchmaking, parties, notifications, capacity requests).
- **Repository Layer**: PostgreSQL for persistent data, Redis for queues and ephemeral state.
- **Orchestrator Layer**: Manages AWS EC2 lifecycles (creates instances when queue is full, terminates when dead).
- **Sidecar Client**: A compiled Go binary injected directly into the CS:S Docker image. It reads the game server's `console.log` and pushes real-time events (e.g., `MATCH_FINISHED`) to the backend via gRPC.

---

## Matchmaking Flow (High-Level)

1. Players enter the matchmaking queue stored in Redis.
2. The `QueReader` safely pops players using atomic Lua scripts.
3. A match is generated in Postgres and an accept window begins.
4. Players accept the match. The final accepting player triggers the server acquisition phase.
5. The orchestrator pulls a `READY` server from the pool (or provisions a new EC2 instance).
6. The backend sends the server IP to the players via WebSockets.
7. The sidecar monitors the game and streams events back to the backend.

---

## Running Locally (Docker Compose)

You can spin up the entire architecture—Database, Cache, Backend, and a fully functional CS:S Dedicated Server with the injected Sidecar—using Docker Compose.

### Prerequisites
- Docker & Docker Compose
- No local Go/Postgres installation required!

### Quick Start

1. Start the stack:
   ```bash
   docker compose up --build
   ```
   *(Note: The first run takes ~5 minutes as it uses SteamCMD to download the ~2.5GB Counter-Strike: Source dedicated server files.)*

2. The services will bind to:
   - **Backend API**: `localhost:8080`
   - **Orchestrator gRPC**: `localhost:6767`
   - **CS:S Game Server**: `27015` (TCP/UDP)
   - **Postgres/Redis**: `5432` / `6379`

The sidecar will automatically connect to the local backend and start streaming heartbeats and game logs.

---

## Testing & CI/CD

- **Unit Tests**: Run locally via `go test -tags=unit ./...`
- **Integration Tests**: Tests against real Postgres and Redis instances.
- **End-to-End (E2E) CI**: On every push to `main`, GitHub Actions automatically:
  1. Boots the full `docker-compose` stack.
  2. Verifies the Backend and Game Server boot successfully.
  3. Asserts that the Sidecar successfully establishes a gRPC stream and sends heartbeats.
- **Automated Deployments**: A GitHub Actions workflow builds the multi-stage Distroless image and pushes it to AWS ECR.

---

## Design Goals

- **Race-Condition Free**: Strict transactional boundaries (`txutil`) and `FOR UPDATE` locks on matches.
- **Cost-Efficient**: EC2 servers are ephemeral, requested on-demand, and terminated via gRPC triggers.
- **Developer Experience**: A single `docker compose up` command mimics the entire production AWS architecture locally.
