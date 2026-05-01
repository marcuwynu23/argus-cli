package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"gopkg.in/yaml.v2"
)

// ==========================
// CONFIG
// ==========================

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

// ==========================
// GLOBAL STATE
// ==========================

var (
	backends      []string
	currentServer uint64

	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	backendHealth = map[string]bool{}
	healthMutex   sync.RWMutex

	healthCheckFreq = 5 * time.Second
	healthTimeout   = 2 * time.Second
)

// ==========================
// CONFIG
// ==========================

func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

func applyEnvOverrides(cfg *Config) {
	if host := os.Getenv("HARIBON_HOST"); host != "" {
		cfg.MainHost = host
	}

	if port := os.Getenv("HARIBON_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.MainPort = p
		}
	}
}

// ==========================
// HEALTH
// ==========================

func isHealthy(b string) bool {
	healthMutex.RLock()
	defer healthMutex.RUnlock()
	return backendHealth[b]
}

func setHealth(b string, status bool) {
	healthMutex.Lock()
	defer healthMutex.Unlock()
	backendHealth[b] = status
}

func startHealthChecks() {
	go func() {
		ticker := time.NewTicker(healthCheckFreq)
		defer ticker.Stop()

		for range ticker.C {
			for _, b := range backends {
				go checkBackend(b)
			}
		}
	}()
}

func checkBackend(b string) {
	ctx, cancel := context.WithTimeout(context.Background(), healthTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b, nil)
	if err != nil {
		setHealth(b, false)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		setHealth(b, false)
		return
	}
	defer resp.Body.Close()

	setHealth(b, resp.StatusCode >= 200 && resp.StatusCode < 400)
}

// ==========================
// LOAD BALANCER
// ==========================

func getNextBackend() (string, error) {
	if len(backends) == 0 {
		return "", fmt.Errorf("no backends configured")
	}

	for range backends {
		i := atomic.AddUint64(&currentServer, 1)
		b := backends[int(i)%len(backends)]

		if isHealthy(b) {
			return b, nil
		}
	}

	return "", fmt.Errorf("no healthy backend available")
}

// SAFE PROXY HANDLER
func loadBalancer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	for range backends {
		server, err := getNextBackend()
		if err != nil {
			http.Error(w, "No healthy backend", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)

		req, err := http.NewRequestWithContext(ctx, r.Method, server, r.Body)
		if err != nil {
			cancel()
			continue
		}

		// forward headers safely
		for k, v := range r.Header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}

		resp, err := httpClient.Do(req)
		cancel()

		if err != nil {
			setHealth(server, false)
			continue
		}
		defer resp.Body.Close()

		// DO NOT forward hop-by-hop / content-length headers
		hopHeaders := map[string]bool{
			"Content-Length":    true,
			"Transfer-Encoding": true,
			"Connection":        true,
		}

		for k, v := range resp.Header {
			if hopHeaders[k] {
				continue
			}
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.WriteHeader(resp.StatusCode)

		_, copyErr := io.Copy(w, resp.Body)
		if copyErr != nil {
			log.Printf("copy error: %v", copyErr)
		}

		log.Printf("%s %s -> %s [%d] (%v)",
			r.Method, r.URL.Path, server, resp.StatusCode, time.Since(start))

		return
	}

	http.Error(w, "All backend servers failed", http.StatusServiceUnavailable)
}

// ==========================
// CONFIG PATH
// ==========================

func resolveConfigPath(cli string) string {
	if cli != "" {
		return cli
	}
	return "./haribon-config.yml"
}

// ==========================
// HELP
// ==========================

func printHelp() {
	fmt.Print(`Haribon Load Balancer

Usage:
  haribon start [options]

Options:
  --config string  Path to config file
  -h, --help       Show help
`)
}

// ==========================
// START
// ==========================

func startCommand(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)

	var configPath string
	fs.StringVar(&configPath, "config", "", "config file path")
	_ = fs.Parse(args)

	config, err := loadConfig(resolveConfigPath(configPath))
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	applyEnvOverrides(&config)

	for _, b := range config.Backends {
		backends = append(backends, b.Host)
		backendHealth[b.Host] = false
	}

	startHealthChecks()

	addr := fmt.Sprintf("%s:%d", config.MainHost, config.MainPort)

	http.HandleFunc("/", loadBalancer)

	log.Printf("running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// ==========================
// MAIN
// ==========================

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "start":
		startCommand(os.Args[2:])
	case "-h", "--help":
		printHelp()
	default:
		fmt.Println("unknown command:", os.Args[1])
		printHelp()
	}
}