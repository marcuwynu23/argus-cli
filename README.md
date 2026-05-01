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

## Features

- Round-robin load balancing
- Simple configuration using YAML
- Logging of request transactions
- Easy to set up and use

## Prerequisites

- Go (version 1.16 or later) installed on your machine
- Basic knowledge of HTTP and RESTful services

## Installation

1. **Clone the repository:**

```bash
git clone https://github.com/marcuwynu23/haribon.git
cd haribon
```

2. **Build the application:**

```bash
go build -o haribon main.go
```

3. **Run the application:**

```bash
./haribon
```

## Configuration

Haribon uses a `harbor-config.yml` file for configuration. Create a file named `harbor-config.yml` in the same directory as the executable with the following structure:

```yml
host: "localhost"
port: 4444
backends:
  - url: "http://localhost:4441"
  - url: "http://localhost:4442"
  - url: "http://localhost:4443"
```

## Configuration Parameters

**host**: The hostname or IP address where the load balancer will listen for incoming requests. Default is `localhost`.

**port**: The port number on which the load balancer will run. Default is `4444`.

**backends**: A list of backend servers to which the load balancer will forward requests. Each backend should have a `url` field specifying the full URL of the backend server.

## Usage

1. Start your backend servers on the specified ports (e.g. `4441`, `4442`, `4443`)

2. Run Haribon:

```bash
./haribon
```

3. Send requests to the load balancer:

```bash
curl http://localhost:4444
```

Haribon forwards requests to backend servers in a round-robin manner.

## Logging

Haribon logs HTTP request transactions to a file named `load_balancer.log`.

## Contributing

Contributions are welcome. Feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
