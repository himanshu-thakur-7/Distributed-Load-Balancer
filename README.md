# Distributed Load Balancer System

This project implements a distributed load balancer system with dynamic backend management, health monitoring, and orchestration. It demonstrates core system design concepts such as service discovery, health checks, and real-time updates using Redis Pub/Sub.

---

## Architecture & Components

### 1. Clients
- Represented as users or services sending HTTP requests to the Load Balancer.
- Not implemented as a separate service; you can use `curl`, Postman, or browser to simulate clients.

### 2. Load Balancer
- **Location:** `backend/load-balancer/`
- **Responsibilities:**
  - Receives HTTP requests from clients (main endpoint: `/process`).
  - Selects a backend using strategies like Round Robin or Least Connections.
  - Forwards requests to healthy backend servers and returns their responses.
  - Loads the initial backend list from Redis at startup.
  - Subscribes to Redis Pub/Sub (`backend_changes` channel) for real-time backend status updates.
  - Maintains an in-memory list of healthy backends.
- **Key Files:** `main.go`, `Dockerfile`

### 3. Backend Servers
- **Location:** `backend/node/`
- **Responsibilities:**
  - Stateless HTTP servers (multiple instances, e.g., Backend1, Backend2, Backend3, Backend4).
  - Each exposes:
    - `/process`: Handles requests with a simulated delay.
    - `/health`: Returns health status (used by orchestrator).
  - Can be killed/restarted independently (Docker containers).
- **Key Files:** `main.go`, `Dockerfile`

### 4. Orchestrator (Control Plane)
- **Location:** `backend/orchestrator/`
- **Responsibilities:**
  - Periodically polls each backend's `/health` endpoint.
  - Determines healthy/unhealthy status.
  - Updates Redis with backend status (only if changed).
  - Publishes events to Redis Pub/Sub (`backend_changes` channel) if status changes.
  - Does **not** handle client traffic or interact directly with the Load Balancer.
- **Key Files:** `main.go`, `Dockerfile`

### 5. Redis (Source of Truth)
- **Location:** Docker service (`redis`)
- **Responsibilities:**
  - Stores backend metadata:
    - Set `backends` for backend IDs.
    - Hashes like `backend:backend1` with fields: `url`, `status`, `last_checked`.
  - Pub/Sub channel `backend_changes` for backend status events (JSON: `{backend_id, status}`).

---

## Folder Structure
```
backend/
├── docker-compose.yml
├── load-balancer/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── node/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── orchestrator/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
└── README.md
Screenshot 2026-02-02 at 12.58.02 AM.png (system design diagram)
```

---

## How It Works

1. **Startup:**
   - Redis is seeded with backend info by `redis-init` job.
   - Orchestrator starts polling `/health` endpoints.
   - Load Balancer loads healthy backends from Redis and subscribes to `backend_changes`.

2. **Request Flow:**
   - Client sends HTTP request to Load Balancer (`/process`).
   - Load Balancer selects a healthy backend and forwards the request.
   - Backend processes and responds; Load Balancer returns response to client.

3. **Health Monitoring:**
   - Orchestrator checks each backend's `/health` endpoint.
   - If a backend's status changes, orchestrator updates Redis and publishes an event.
   - Load Balancer updates its in-memory list in real time via Pub/Sub.

4. **Dynamic Scaling:**
   - Backends can be killed or restarted independently.
   - System adapts automatically to backend health changes.

---

## Running the System

1. **Build & Start All Services:**
   ```sh
   cd backend
   docker compose up --build
   ```

2. **Send Requests:**
   ```sh
   curl http://localhost:8080/process
   ```

3. **Simulate Backend Failure:**
   ```sh
   docker stop backend2
   # or toggle health endpoint if implemented
   docker exec backend2 sh -lc 'wget -qO- http://localhost:8080/toggle-health'
   ```

4. **Check Redis State:**
   ```sh
   docker exec -it redis redis-cli
   SMEMBERS backends
   HGETALL backend:backend1
   SUBSCRIBE backend_changes
   ```

---

## System Design Diagram

The following diagram illustrates the architecture and interactions between components:

![System Design Diagram](Screenshot%202026-02-02%20at%2012.58.02%E2%80%AFAM.png)

---

## Notes
- All components are containerized using Docker.
- Redis is the single source of truth for backend state.
- The orchestrator is stateless and can be scaled independently.
- The system is resilient to backend failures and supports dynamic backend registration/removal.

---

## License
MIT
