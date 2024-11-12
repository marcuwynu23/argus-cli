# ARGUS - Load Balancer

ARGUS is a simple and efficient load balancer written in Go. It distributes incoming HTTP requests across multiple backend servers using a round-robin algorithm, ensuring optimal resource utilization and improved performance.

## Features

- Round-robin load balancing
- Simple configuration using YAML
- Logging of request transactions
- Easy to set up and use

## Prerequisites

- Go (version 1.16 or later) installed on your machine.
- Basic knowledge of HTTP and RESTful services.

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/yourusername/argus.git
   cd argus
   ```
2. **Build the application:**
  ```bash
  go build -o argus main.go
  ```

3. **Run the application:**
  ```bash
  ./argus
  ```

## Configuration

ARGUS uses a config.yml file for configuration. Create a file named config.yml in the same directory as the executable with the following structure:
```yml
host: "localhost"
port: 4444
backends:
  - url: "http://localhost:4441"
  - url: "http://localhost:4442"
  - url: "http://localhost:4443"
```

### Configuration Parameters
**host**: The hostname or IP address where the load balancer will listen for incoming requests. Default is "localhost".

**port**: The port number on which the load balancer will run. Default is 4444.

**backends**: A list of backend servers to which the load balancer will forward requests. Each backend should have a url field specifying the full URL of the backend server.


### Example `config.yml`

```yml
host: "localhost"
port: 4444
backends:
  - url: "http://localhost:4441"
  - url: "http://localhost:4442"
  - url: "http://localhost:4443"
```


## Usage

1. Start your backend servers on the specified ports (e.g., 4441, 4442, 4443).

2. Run the ARGUS load balancer:
  
  ```bash
  ./argus
  ```

3. Send requests to the load balancer:

You can use curl or any HTTP client to send requests to the load balancer:

```bash
curl http://localhost:4444
```
The load balancer will forward the requests to the backend servers in a round-robin manner.

## Logging
ARGUS logs all HTTP request transactions to a file named load_balancer.log. You can check this file for details about the requests processed by the load balancer.

## Contributing
Contributions are welcome! If you have suggestions or improvements, feel free to open an issue or submit a pull request.

## License
This project is licensed under the MIT License. See the LICENSE file for details.