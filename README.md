<div align="center">
  <h1> Haribon </h1>
</div>

<p align="center">
  <img src="https://img.shields.io/github/stars/marcuwynu23/haribon.svg" alt="Stars Badge"/>
  <img src="https://img.shields.io/github/forks/marcuwynu23/haribon.svg" alt="Forks Badge"/>
  <img src="https://img.shields.io/github/issues/marcuwynu23/haribon.svg" alt="Issues Badge"/>
  <img src="https://img.shields.io/github/license/marcuwynu23/haribon.svg" alt="License Badge"/>
</p>

Haribon is a lightweight Go-based load balancer designed for simplicity, observability, and production readiness.
It supports round-robin routing, health-aware balancing, structured logging, and Loki/Promtail integration.

---

## Features

- Round-robin load balancing
- Health-aware routing (skip unhealthy backends)
- Structured JSON logging (Loki-ready)
- Promtail-compatible log output
- Environment variable overrides
- Safe HTTP reverse proxying
- Automatic fallback logging (stdout if file fails)
- Configurable log file creation and directory auto-creation

---

## Installation

```bash
git clone https://github.com/marcuwynu23/haribon.git
cd haribon
go build -o haribon main.go
```

---

## Configuration

### haribon-config.yml

```yaml
host: "0.0.0.0"
port: 4444

logging: true

backends:
  - url: "http://localhost:4441"
  - url: "http://localhost:4442"
  - url: "http://localhost:4443"
```

---

## Environment Overrides

| Variable     | Description    | Example |
| ------------ | -------------- | ------- |
| HARIBON_HOST | Bind host      | 0.0.0.0 |
| HARIBON_PORT | Listening port | 4444    |

---

## Run

```bash
./haribon start --config haribon-config.yml
```

---

## Load Balancing Behavior

Haribon uses:

- Round-robin selection
- Health-aware backend selection
- Automatic fallback if no healthy backend is available

If health data is empty (startup/testing mode), all backends are treated as healthy.

---

## Health Checking

- Backend health state stored in memory
- Unhealthy backends are skipped automatically
- Only healthy services receive traffic

---

## Logging

Haribon outputs structured JSON logs compatible with Loki and Promtail.

### Example log

```json
{
  "time": "2026-05-04T01:31:55.3551454Z",
  "method": "GET",
  "path": "/",
  "backend": "http://localhost:4442",
  "status": 200,
  "duration_ms": 5,
  "level": "info"
}
```

### Logging behavior

- Logs are written in JSON format
- Supports stdout and file output simultaneously
- Automatically creates log file if missing
- Falls back to stdout if file cannot be created

---

## Docker Usage

### Pull image

```bash
docker pull ghcr.io/marcuwynu23/haribon:latest
```

### Run

```bash
docker run -d -p 4444:4444 ghcr.io/marcuwynu23/haribon:latest
```

---

## Recommended Structure

```
data/
  haribon-config.yml
  haribon.log
```

---

## Example Request

```bash
curl http://localhost:4444
```

Requests are distributed using round-robin scheduling with health filtering.

---

## Observability Stack

Haribon is designed to integrate with:

- Grafana
- Loki
- Promtail

Logs are structured for direct ingestion.

---

## Testing

```bash
go test ./...
```

---

## Architecture Notes

- Stateless proxy core
- Atomic counter for routing
- Mutex-protected log writer
- RWMutex backend health store
- Context-based request cancellation

---

## Roadmap

- Active health check scheduler
- Retry policy per backend
- Circuit breaker
- Metrics endpoint (/metrics)
- Prometheus integration
- Weighted load balancing

---

## License

MIT License
