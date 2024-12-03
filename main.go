// load_balancer.go
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	MainHost string    `yaml:"host"`
	MainPort int       `yaml:"port"`
	Logging  bool      `yaml:"logging"`
	LogPath  string    `yaml:"log_path"`
	Backends []Backend `yaml:"backends"`
}

type Backend struct {
	Host string `yaml:"url"`
}

var (
	backends      []string
	currentServer uint64
	logFile       *os.File
)

func logRequest(r *http.Request, server string, statusCode int) {
	if logFile != nil {
		log.SetOutput(logFile)
	} else {
		log.SetOutput(os.Stdout)
	}
	log.Printf("Request: %s %s, Forwarded to: %s, Status Code: %d\n", r.Method, r.URL, server, statusCode)
}

func loadBalancer(w http.ResponseWriter, r *http.Request) {
	// Track the number of servers that failed
	failures := 0
	var resp *http.Response
	var err error
	var server string

	// Loop through all backend servers until one responds successfully
	for {
		// Use atomic operations to get the next server in a round-robin manner
		nextServer := atomic.AddUint64(&currentServer, 1) % uint64(len(backends))
		server = backends[nextServer]

		// Forward the request to the selected backend server
		startTime := time.Now() // Start time for logging duration
		resp, err = http.Get(server)
		if err == nil {
			// If the request is successful, break out of the loop
			defer resp.Body.Close()
			// Copy the response from the backend server to the client
			w.WriteHeader(resp.StatusCode)
			_, err = io.Copy(w, resp.Body) // Correctly copy the response body
			if err != nil {
				http.Error(w, "Error writing response", http.StatusInternalServerError)
				logRequest(r, server, http.StatusInternalServerError) // Log error
				return
			}

			// Log the successful request
			logRequest(r, server, resp.StatusCode)

			// Log the duration of the request
			duration := time.Since(startTime)
			log.Printf("Request processed in %v\n", duration)
			return
		}

		// If we reach here, it means the request to the current server failed
		failures++

		// If all servers fail, return an error
		if failures == len(backends) {
			http.Error(w, "All backend servers are down", http.StatusServiceUnavailable)
			logRequest(r, server, http.StatusServiceUnavailable) // Log error for the last tried server
			return
		}

		// Otherwise, continue to the next server in the round-robin cycle
	}
}

func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	return config, err
}

func main() {
	// Load configuration from YAML file
	configFile := "argus-config.yml"
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Set backend servers
	for _, backend := range config.Backends {
		backends = append(backends, backend.Host)
	}
	// Set up logging if enabled
	if config.Logging {
		if config.LogPath == "" {
			// Set default log path to the current working directory
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			config.LogPath = filepath.Join(currentDir, "argus.log")
		}

		var err error
		logFile, err = os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()
	}
	// Start the load balancer
	http.HandleFunc("/", loadBalancer)
	fmt.Printf("Load balancer is running on %s:%d...\n", config.MainHost, config.MainPort)
	address := fmt.Sprintf("%s:%d", config.MainHost, config.MainPort)
	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
