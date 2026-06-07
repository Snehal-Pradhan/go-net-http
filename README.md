
<div>
  <img alt="Project Status" src="https://img.shields.io/badge/Project%20Status-Completed%20/%20Stable-success">
</div>

<h3 align="center">
  Go net/http Server
  <br/>
  with Middleware Chaining and TLS
</h3>
<div align="center">
  <img alt="Go logo" src="public/logo.png" width="20%">
</div>

<p align="center">
    <a href="#features">Features</a> ·
    <a href="#routes">Routes</a> ·
    <a href="#quick-start">Quick Start</a> ·
    <a href="#testing">Testing</a>
</p>


&nbsp;

> **Project Status:** 🟢 **Complete & Stable.** All core features, architectural benchmarks, and test suites are fully implemented and verified. This repository is in maintenance mode.

A production-style HTTPS server built entirely on Go's standard `net/http` package — middleware chaining, Bearer token auth, subrouting, cookies, path params, host-based routing, and a CRUD todo API — without any third-party frameworks.

**Tech stack:** Go · `net/http` · `encoding/json` · `sync`

## Features

| Routing                                                 | Middleware                                              | Auth & Security                                          | Storage & TLS                                      |
| :------------------------------------------------------ | :------------------------------------------------------ | :------------------------------------------------------- | :------------------------------------------------- |
| Path parameters via `r.PathValue("id")`                 | Logging — method, path, duration                        | Bearer token extraction from `Authorization` header      | In-memory thread-safe store (`sync.RWMutex`)       |
| Method-based routing (`GET /path`, `POST /path`)        | Custom header injection (`X-Custom: hello`)             | Token propagation via `context.WithValue`                | JSON serialization via `encoding/json`             |
| Host-based routing (`api.example.com/hello`)            | Composible middleware stack (`createStack`)             | Protection on POST/PUT/DELETE todo endpoints             | TLS 1.3 via `ListenAndServeTLS` on `:443`          |
| Subrouting with `http.StripPrefix` (`/users/`, `/todos/`)| Response status code capture via wrapped `ResponseWriter`| 401 on missing token                                     | Self-signed certs via OpenSSL                      |

## Routes

| Method   | Path                    | Auth | Response                                |
| :------- | :---------------------- | :--- | :-------------------------------------- |
| `GET`    | `/`                     | No   | `Home`                                  |
| `GET`    | `/{id}`                 | Yes  | `{"id":"42","token":"..."}`             |
| `GET`    | `/set-cookie`           | No   | `cookie set`                            |
| `GET`    | `/get-cookie`           | No   | `cookie: abc123`                        |
| `GET`    | `/users/`               | No   | `list users`                            |
| `POST`   | `/users/`               | No   | `create user`                           |
| `GET`    | `/users/{id}`           | No   | `{"id":"123"}`                          |
| `GET`    | `/todos/`               | No   | `[{"id":1,"title":"...","completed":false}]` |
| `POST`   | `/todos/`               | Yes  | `{"id":1,"title":"...","completed":false}`  |
| `GET`    | `/todos/{id}`           | No   | `{"id":1,"title":"...","completed":false}`  |
| `PUT`    | `/todos/{id}`           | Yes  | `{"id":1,"title":"...","completed":true}`   |
| `DELETE` | `/todos/{id}`           | Yes  | `204 No Content`                         |
| `GET`    | `api.example.com/hello` | No   | `hello from api subdomain`              |

## Quick Start

```bash
# Generate TLS certs (one-time)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key -out server.crt \
  -subj "/CN=localhost"

# Start server
go run .

# Create a todo (authenticated)
curl -sk -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-token" \
  -d '{"title":"learn Go"}' \
  https://localhost/todos/

# List todos
curl -sk https://localhost/todos/

# Protected route without token (returns 401)
curl -sk https://localhost/42
```

## Testing

```bash
go test -v .
```

| File             | Tests | What it verifies                                                          |
| :--------------- | :---- | :------------------------------------------------------------------------ |
| `main_test.go`   | 13    | Middleware order, status capture, auth extraction, cookies, user router   |
| `todo_test.go`   | 15    | CRUD operations, auth enforcement, not found, invalid input, validation   |

## Future Scope

- Persistent storage (PostgreSQL or SQLite)
- Rate limiting middleware
- Request ID tracing middleware
- OpenAPI / Swagger documentation
- Docker containerization
- CI pipeline with automated tests

<br>

<div align="center">
  <p>Built with Go</p>
</div>

&nbsp;
