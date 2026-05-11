# Net-Centric Programming Group Project: MangaHub - Manga & Comic Tracking System

| Student name     | Student ID  |
| :--------------- | :---------- |
| Nguyen Danh Huy  | ITITIU22071 |
| Nguyen Minh Khoi | ITITIU23013 |

### 1. Objectives

- Allow students to gain practical experience of network application development using Go with realistic scope
- Experience all five required communication protocols (TCP, UDP, HTTP, gRPC, WebSocket) through hands-on implementation
   Strengthen understanding of networking concepts through manageable, progressive implementation
- Develop foundational skills in concurrent programming and basic distributed system patterns
- Create a working system that demonstrates network programming competency within academic constraints

### Running Test Instruction
---

## Prerequisites

- Go installed and project dependencies resolved
- [`wscat`](https://github.com/websockets/wscat) installed (`npm install -g wscat`)
- Valid JWT token for `testuser2@mangahub.com`
- Six terminal windows ready

---

## Step 1: Start the Core Server

**Terminal 1**

```bash
go run cmd/api-server/main.go
```

> Wait until the server confirms it is listening on ports **8080**, **9090**, **9091**, and **50051** before proceeding.

---

## Step 2: Initialize Protocol Listeners

Open three separate terminals and start each listener.

**Terminal 2 — TCP Sync Listener**

```bash
go run cmd/test_tcp/main.go
```

**Terminal 3 — UDP Notification Listener**

```bash
go run cmd/mangahub/main.go notify subscribe
```

**Terminal 4 — WebSocket Chat Client**

```bash
wscat -c ws://localhost:8080/chat
```

---

## Step 3: Run the Integration Action Sequence

Open a **5th terminal** and run the following HTTP commands in order.

### Action A — Authentication (HTTP)

Verify user login. A successful response returns a token matching the one used in Actions B and C.

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "testuser2@mangahub.com", "password": "supersecretpassword"}'
```

---

### Action B — Reading Progress Update (HTTP → TCP Broadcast)

Send a chapter progress update using your JWT token.

```bash
curl -X PUT http://localhost:8080/users/progress \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"manga_id": "one-piece", "chapter": 1110}'
```

✅ **Verification:** Switch to **Terminal 2**. You should see a `[SYNC RECEIVED]` message containing the updated chapter data.

---

### Action C — Chapter Release Notification (HTTP → UDP Broadcast)

Trigger a global chapter release notification.

```bash
curl -X POST http://localhost:8080/users/notify \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"manga_id": "one-piece", "message": "Chapter 1111 is OUT!"}'
```

✅ **Verification:** Switch to **Terminal 3**. You should see the UDP release alert appear.

---

### Action D — Community Discussion (WebSocket)

In **Terminal 4** (the active `wscat` session), type the following message and press **Enter**:

```json
{"username": "testuser2", "message": "Did you see that notification? Chapter 1111 is crazy!"}
```

✅ **Verification:** The message echoes back in Terminal 4. If a second `wscat` client is open, the message appears there too.

---

## Step 4: Internal Data Audit (gRPC)

Confirm the gRPC service is running and reading from the same database as the HTTP server.

**Terminal 6**

```bash
go run cmd/test_grpc/main.go
```

✅ **Expected output:**

```
✅ gRPC Response Received! Title: One Piece...
```

---

## Summary of Expected Results

| Action | Protocol | Terminal to Watch | Expected Result |
|--------|----------|-------------------|-----------------|
| A — Login | HTTP | Terminal 5 | JSON response with auth token |
| B — Progress Update | HTTP → TCP | Terminal 2 | `[SYNC RECEIVED]` with chapter data |
| C — Chapter Notify | HTTP → UDP | Terminal 3 | UDP release alert |
| D — Chat Message | WebSocket | Terminal 4 | Message echo (+ other clients) |
| gRPC Audit | gRPC | Terminal 6 | `✅ gRPC Response Received! Title: One Piece...` |

