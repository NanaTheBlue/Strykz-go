# Go Matchmaking Backend API

This repository contains a Go-based backend API that powers a real-time multiplayer matchmaking and social system inspired by faceit.

The backend is designed with clear internal layering (API, services, repositories), focusing on real-time communication and correctness. It currently uses PostgreSQL, Redis, and WebSockets.

---

## Features

### Authentication & Sessions

- Custom authentication system written in Go
- PostgreSQL-backed refresh token storage
- Secure login and registration flows

### Matchmaking System

- Fifo based matchmaking
- Support for solo players and parties (soon)
- Atomic match creation to prevent race conditions
- Timeout handling for players who do not accept matches (soon)
- **Match acceptance / decline flow (in progress)**

### Party System

- Party creation and management
- Party leader and member roles
- Party invite system

### Real-Time Communication

- WebSocket-based notifications
- Online user tracking (Hub)
- Party invites and matchmaking updates

### Messaging & Queues

- Redis used as the primary matchmaking queue
- Atomic queue operations to prevent race conditions
- re-queueing handled via Redis

### Social Features

- Notifications service
- Online presence tracking
- Extensible foundation for friends and social interactions

---

## Tech Stack

- **Language:** Go (Golang)
- **Database:** PostgreSQL
- **Cache / PubSub:** Redis
- **Queue System:** Redis
- **Real-Time:** WebSockets
- **Frontend Integration:** Designed to be consumed by a Next.js BFF / frontend

---

## Architecture Overview

The backend is implemented as a **single Go application** with clear internal separation of concerns:

- **API Layer**

  - HTTP endpoints for authentication, matchmaking, and social features

- **Service Layer**

  - Encapsulates business logic (matchmaking, parties, notifications)
  - Coordinates between PostgreSQL, Redis, and WebSockets

- **Repository Layer**

  - PostgreSQL repositories for persistent data
  - Redis for matchmaking queues and ephemeral state

- **Real-Time Layer**

  - WebSocket hub for pushing notifications and matchmaking events

This structure keeps the codebase modular.

---

## Matchmaking Flow (High-Level)

1. Player or party enters the matchmaking queue stored in Redis
2. Players are grouped by MMR bucket
3. Atomic Redis operations ensure safe dequeuing and match creation
4. When a valid match is found:

   - A unique match ID is generated
   - Players are notified via WebSocket
   - Acceptance handling is planned but not yet finalized

5. If a match cannot be completed:

   - Players are safely returned to the queue

---

## API Overview

### Authentication

- `POST /register`
- `POST /login`
- `GET /renew`

### Matchmaking

- `POST /matchmaking/queue`
- `POST /matchmaking/dequeue` (Future)
- `POST /matchmaking/accept` (Future)
- `POST /matchmaking/decline` (Future)

### Parties (Future Routes, Service is already made)

- `POST /parties/create`
- `POST /parties/invite`
- `POST /parties/accept`
- `POST /parties/leave`

### Social / Notifications

- `GET /notification/` (WebSocket connection)

---

## Data Storage

- **PostgreSQL**

  - Users
  - Sessions
  - Parties and party members
  - Match records

- **Redis**

  - Matchmaking queues

  - Online users (Future)

  - Pub/Sub for notifications

  - Matchmaking queue

---

## Running Locally

### Prerequisites

- Go 1.25.3+
- PostgreSQL
- Redis

### Environment Variables

```env
POSTGRES_URL=
REDIS_ADDRESS=
REDIS_PASSWORD=
JWT_SECRET=
TEST_POSTGRES_URL=
TEST_REDIS_ADDRESS=
TEST_REDIS_PASSWORD=
```

## Testing

- Unit tests for services and repositories
- Integration tests for Redis, PostgreSQL, and WebSocket flows
- Test utilities for creating users and cleaning up state

---

## Design Goals

- Predictable and debuggable matchmaking logic
- Clear separation of concerns
- Real-time responsiveness
- Safe concurrency and atomic operations
- Extensible foundation for future features

---

## Future Improvements

- Match history and statistics
- Ranked seasons
- Horizontal scaling for matchmaking workers
