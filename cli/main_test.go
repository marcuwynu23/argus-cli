package main

import "testing"

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