package main

import (
	"strings"
	"testing"
	"os"
	"sync/atomic"
)

func TestHealthMap(t *testing.T) {
	setHealth("http://a", true)
	if !isHealthy("http://a") {
		t.Fatal("expected healthy backend")
	}
}

func TestRoundRobinSkipsDead(t *testing.T) {
	backends = []string{"a", "b"}

	setHealth("a", false)
	setHealth("b", true)

	b, err := getNextBackend()
	if err != nil {
		t.Fatal(err)
	}

	if b != "b" {
		t.Fatalf("expected b, got %s", b)
	}
}

func TestNoBackends(t *testing.T) {
	backends = []string{}
	_, err := getNextBackend()
	if err == nil {
		t.Fatal("expected error")
	}
}

func resetState() {
	backends = nil
	atomic.StoreUint64(&currentServer, 0)

	backendHealth = map[string]bool{}
}



func TestGetNextBackend_RoundRobin(t *testing.T) {
	resetState()
	backends = []string{"a", "b", "c"}
	currentServer = 0

	first, _ := getNextBackend()
	second, _ := getNextBackend()
	third, _ := getNextBackend()

	if first != "a" || second != "b" || third != "c" {
		t.Fatalf("round robin failed: %s %s %s", first, second, third)
	}
}

func TestWriteLog_JSONFormat(t *testing.T) {
	var buf strings.Builder
	logWriter = &buf

	writeLog(LogEntry{
		Method:  "GET",
		Path:    "/",
		Backend: "http://localhost",
		Status:  200,
		Level:   "info",
	})

	out := buf.String()

	if !strings.Contains(out, `"method":"GET"`) {
		t.Fatalf("missing method field in log: %s", out)
	}

	if !strings.Contains(out, `"backend"`) {
		t.Fatalf("missing backend field in log")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	os.Setenv("HARIBON_HOST", "127.0.0.1")
	os.Setenv("HARIBON_PORT", "9999")

	cfg := Config{}
	applyEnvOverrides(&cfg)

	if cfg.MainHost != "127.0.0.1" {
		t.Fatal("host override failed")
	}
	if cfg.MainPort != 9999 {
		t.Fatal("port override failed")
	}
}