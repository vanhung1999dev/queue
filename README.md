# gstream

A lightweight in-memory message broker written in Go. Supports HTTP-based message publishing and consuming with topic support, offset tracking, message retention, and periodic cleanup.

## 📦 Features

- HTTP-based **Producers** and **Consumers**
- In-memory **Message Storage**
- **Topic-based** message handling
- **Offset-based** consumption
- **Retention policy** for automatic expiration
- Periodic **cleanup of expired messages**
- Configurable via environment variables (from `.env` or shell)

## 🚀 Quick Start

### Requirements

- Go 1.18+

### 1. Build

```bash
make build
```

This creates a binary at `./bin/gstream`.

### 2. Configure

Create a `.env` file in the project root:

```
RETENTION_DURATION=5m
CLEANUP_INTERVAL=1m
```

Or export them in your shell:

```bash
export RETENTION_DURATION=5m
export CLEANUP_INTERVAL=1m
```

### 3. Run the server

```bash
make run
```

Server starts on port `:3000`.

## 📡 API Reference

### ✅ Publish a Message

**POST /publish/<topic>**

- Body: raw bytes (message payload)

```bash
curl -X POST http://localhost:3000/publish/news -d "Hello world!"
```

- `202 Accepted`: Message stored
- `400 Bad Request`: Invalid request path

### 📥 Consume a Message

**GET /consume/<topic>/<offset>**

- Returns message at given offset (if not expired)

```bash
curl http://localhost:3000/consume/news/0
```

- `200 OK`: Message data
- `204 No Content`: Message expired or not found
- `400 Bad Request`: Offset parsing error

## 🧠 How It Works

- Messages are stored in-memory per topic
- Each message has a timestamp
- Messages are removed if older than `RETENTION_DURATION`
- A cleaner runs every `CLEANUP_INTERVAL` to purge old messages
- Topic stores are initialized lazily on first publish or consume

## 🗂️ Project Structure

| File         | Purpose                            |
|--------------|------------------------------------|
| `main.go`    | Entry point, loads env, starts server |
| `producer.go`| HTTP POST and GET handlers         |
| `server.go`  | Server + store/topic logic         |
| `storage.go` | In-memory store w/ retention logic |
| `consumer.go`| (Stub) for future WebSocket/stream support |
| `Makefile`   | Build and run targets              |

## 🧪 Example Test

```bash
# Publish a message to 'test' topic
curl -X POST http://localhost:3000/publish/test -d "test message"

# Read message at offset 0
curl http://localhost:3000/consume/test/0
```

## 🔮 Future Improvements

- WebSocket support for real-time consumers
- Persistent storage (e.g. BoltDB, Redis)
- Auth & rate limiting
- Observability: metrics/logging enhancements

## 👨‍💻 Author

vanhung — [vanhung1999dev@gmail.com]

## 📄 License

MIT
