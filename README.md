# Rate Limiter Demo (Go + Redis)

This project demonstrates different **rate limiting algorithms** implemented in Go.  
It uses **Redis** as the distributed store for counters/tokens, with a simple backend service protected by a proxy that enforces rate limiting.

---

## Features
- Fixed Window Counter (basic)
- Token Bucket (burst-friendly)
- Redis-backed distributed rate limiting
- Example HTTP proxy that enforces limits before requests hit the backend

---

## Steps

### 1. Run Redis

```
$ docker run -d --name redis -p 6379:6379 redis/redis-stack-server:latest
```

### 2. Run the backend service

```
$ go run backend/main.go
Dummy backend running on :9000
```

### 3. Run the proxy with rate limiter

```
$ go run proxy/main.go
Rate limiter proxy listening on :8080
```

### 4. Test with curl

```
$ for i in {1..10}; do curl -s http://localhost:8080/; echo; done
Hello from backend!
Hello from backend!
Hello from backend!
Hello from backend!
Hello from backend!
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
Too Many Requests
```

## Acknowledgment
This project was developed with help of the OpenAI ChatGPT-5 model.
