<div align="center">
  <h1> Haribon </h1>
</div>

<p align="center">
  <img src="https://img.shields.io/github/stars/marcuwynu23/haribon.svg" alt="Stars Badge"/>
  <img src="https://img.shields.io/github/forks/marcuwynu23/haribon.svg" alt="Forks Badge"/>
  <img src="https://img.shields.io/github/issues/marcuwynu23/haribon.svg" alt="Issues Badge"/>
  <img src="https://img.shields.io/github/license/marcuwynu23/haribon.svg" alt="License Badge"/>
</p>

Haribon is a simple and efficient load balancer written in Go. It distributes incoming HTTP requests across multiple backend servers using a round-robin algorithm, ensuring optimal resource utilization and improved performance.

---

## Features

- Round-robin load balancing
- YAML-based configuration
- Request logging
- Lightweight and fast
- Docker-ready deployment

---

## Prerequisites

- Go 1.23+
- Docker (optional)
- Docker Compose (recommended)

---

## Installation (Local)

```bash
git clone https://github.com/marcuwynu23/haribon.git
cd haribon
go build -o haribon main.go
./haribon
```

---

## Configuration

Haribon uses a `harbor-config.yml` file:

```yml
host: "0.0.0.0"
port: 4444

backends:
  - url: "http://localhost:4441"
  - url: "http://localhost:4442"
  - url: "http://localhost:4443"
```

---

## Logging

Logs are written to:

```
/tmp/haribon.log
```

When running in Docker, ensure this path is mounted for persistence.

---

# Docker Usage

## Pull Image (GHCR)

```bash
docker pull ghcr.io/marcuwynu23/haribon:latest
```

## Run Container

```bash
docker run -d \
  -p 4444:4444 \
  -v $(pwd)/data:/data \
  ghcr.io/marcuwynu23/haribon:latest
```

---

## Environment Variables

| Variable       | Description      | Default                 |
| -------------- | ---------------- | ----------------------- |
| HARIBON_CONFIG | Config file path | /data/harbor-config.yml |
| HARIBON_PORT   | Listening port   | 4444                    |
| HARIBON_LOG    | Log file path    | /data/logs/harbor.log   |

---

## Docker Compose

Example files are located in:

```
./docker-compose/
```

### Start services

```bash
docker compose up -d
```

This will:

- build or pull required images
- start Haribon in detached mode
- expose service on port 4444

---

### Stop services

```bash
docker compose down
```

Stops and removes containers while keeping data intact.

---

### Full cleanup (including volumes)

```bash
docker compose down -v
```

Removes containers and all volumes (⚠ data will be deleted).

---

## Folder Structure (Recommended)

Create this structure for Docker/local persistence:

```
data/
  harbor-config.yml
  logs/
    harbor.log
```

### Setup

```bash
mkdir -p data/logs
touch data/logs/harbor.log
```

### Copy config

```bash
cp docker-compose/examples/harbor-config.yml.example data/harbor-config.yml
```

---

## Usage

Start backend services:

```bash
curl http://localhost:4441
curl http://localhost:4442
curl http://localhost:4443
```

Run Haribon:

```bash
curl http://localhost:4444
```

Requests are distributed using round-robin.

---

## Contributing

Pull requests are welcome. Open issues for bugs or improvements.

---

## License

MIT License
