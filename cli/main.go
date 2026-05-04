package main

import (
	"context"
	"encoding/json"
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
	"path/filepath"
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
// LOG STRUCT (LOKI FRIENDLY)
// ==========================

type LogEntry struct {
	Time       string `json:"time"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Backend    string `json:"backend"`
	Status     int    `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	Level      string `json:"level"`
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

	logWriter io.Writer = os.Stdout
	mu        sync.Mutex
)


var (
	backendHealth = map[string]bool{}
	healthMutex   sync.RWMutex
)

func setHealth(b string, status bool) {
	healthMutex.Lock()
	defer healthMutex.Unlock()
	backendHealth[b] = status
}

func isHealthy(b string) bool {
	healthMutex.RLock()
	defer healthMutex.RUnlock()
	return backendHealth[b]
}


// ==========================
// CONFIG LOADING
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
// LOKI LOGGING
// ==========================

func writeLog(entry LogEntry) {
	entry.Time = time.Now().UTC().Format(time.RFC3339Nano)

	b, err := json.Marshal(entry)
	if err != nil {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	_, _ = logWriter.Write(append(b, '\n'))
}

// ==========================
// LOAD BALANCER CORE
// ==========================
func getNextBackend() (string, error) {
	if len(backends) == 0 {
		return "", fmt.Errorf("no backends configured")
	}

	n := len(backends)
	start := atomic.AddUint64(&currentServer, 1) - 1

	for i := 0; i < n; i++ {
		idx := (int(start) + i) % n
		b := backends[idx]

		// if health map is empty → all healthy (test mode)
		if len(backendHealth) == 0 || isHealthy(b) {
			return b, nil
		}
	}

	return "", fmt.Errorf("no healthy backend available")
}
// ==========================
// HANDLER
// ==========================

func loadBalancer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	for range backends {
		server, err := getNextBackend()
		if err != nil {
			http.Error(w, "No backend available", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)

		req, err := http.NewRequestWithContext(ctx, r.Method, server, r.Body)
		if err != nil {
			cancel()
			continue
		}

		resp, err := httpClient.Do(req)
		cancel()

		if err != nil {
			continue
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)

		writeLog(LogEntry{
			Method:     r.Method,
			Path:       r.URL.Path,
			Backend:    server,
			Status:     resp.StatusCode,
			DurationMS: time.Since(start).Milliseconds(),
			Level:      "info",
		})

		return
	}

	writeLog(LogEntry{
		Method:     r.Method,
		Path:       r.URL.Path,
		Backend:    "",
		Status:     503,
		DurationMS: time.Since(start).Milliseconds(),
		Level:      "error",
	})

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

	if len(config.Backends) == 0 {
		log.Fatal("no backends defined")
	}

	for _, b := range config.Backends {
		backends = append(backends, b.Host)
	}

	if config.Logging {
		// default log path fallback
		if config.LogPath == "" {
			config.LogPath = "./haribon.log"
		}

		// ensure directory exists (important for Linux paths like /var/log)
		if err := os.MkdirAll(filepath.Dir(config.LogPath), 0755); err != nil {
			log.Printf("log dir create warning: %v", err)
		}

		f, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("log file error (fallback to stdout only): %v", err)
			logWriter = os.Stdout
		} else {
			logWriter = io.MultiWriter(os.Stdout, f)
		}
	} else {
		logWriter = os.Stdout
	}

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
		fmt.Println("usage: start --config <file>")
		return
	}

	switch os.Args[1] {
	case "start":
		startCommand(os.Args[2:])
	default:
		fmt.Println("unknown command")
	}
}