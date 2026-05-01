package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

func logRequest(r *http.Request, server string, statusCode int) {
	log.Printf("%s %s -> %s [%d]\n", r.Method, r.URL.Path, server, statusCode)
}

func getNextBackend() (string, error) {
	if len(backends) == 0 {
		return "", fmt.Errorf("no backends configured")
	}
	i := atomic.AddUint64(&currentServer, 1)
	return backends[int(i)%len(backends)], nil
}

func loadBalancer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	for range backends {
		server, err := getNextBackend()
		if err != nil {
			http.Error(w, "No backend available", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, r.Method, server, r.Body)
		if err != nil {
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)

		logRequest(r, server, resp.StatusCode)
		log.Printf("duration=%v\n", time.Since(start))
		return
	}

	http.Error(w, "All backend servers failed", http.StatusServiceUnavailable)
}

func resolveConfigPath(cli string) string {
	if cli != "" {
		return cli
	}
	return "./haribon-config.yml"
}

func printHelp() {
	fmt.Println(`Haribon Load Balancer

Usage:
  haribon start [options]

Commands:
  start            Start the load balancer

Options:
  --config string  Path to config file (default: ./haribon-config.yml)
  -h, --help       Show help
`)
}

func startCommand(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)

	var configPath string
	fs.StringVar(&configPath, "config", "", "config file path")

	_ = fs.Parse(args)

	configFile := resolveConfigPath(configPath)

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if len(config.Backends) == 0 {
		log.Fatal("no backends defined")
	}

	for _, b := range config.Backends {
		backends = append(backends, b.Host)
	}

	if config.Logging {
		if config.LogPath == "" {
			if runtime.GOOS == "linux" {
				config.LogPath = "/var/log/haribon.log"
			} else {
				dir, _ := os.Getwd()
				config.LogPath = filepath.Join(dir, "haribon.log")
			}
		}

		f, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("log file error: %v", err)
		}
		defer f.Close()

		log.SetOutput(io.MultiWriter(f, os.Stdout))
	} else {
		log.SetOutput(os.Stdout)
	}

	addr := fmt.Sprintf("%s:%d", config.MainHost, config.MainPort)
	http.HandleFunc("/", loadBalancer)

	log.Printf("running on %s\n", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
}

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