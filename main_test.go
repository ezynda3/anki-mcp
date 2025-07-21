package main

import (
	"testing"
)

func TestNewAnkiMCPServer(t *testing.T) {
	server := NewAnkiMCPServer()
	if server == nil {
		t.Fatal("NewAnkiMCPServer returned nil")
	}
	if server.ankiClient == nil {
		t.Fatal("AnkiConnect client not initialized")
	}
}

func TestNewAnkiConnect(t *testing.T) {
	client := NewAnkiConnect()
	if client == nil {
		t.Fatal("NewAnkiConnect returned nil")
	}
	if client.URL != defaultAnkiConnectURL {
		t.Errorf("Expected URL %s, got %s", defaultAnkiConnectURL, client.URL)
	}
	if client.Version != ankiConnectVersion {
		t.Errorf("Expected version %d, got %d", ankiConnectVersion, client.Version)
	}
}

func TestNewAnkiConnectWithURL(t *testing.T) {
	customURL := "http://localhost:9999"
	client := NewAnkiConnectWithURL(customURL)
	if client == nil {
		t.Fatal("NewAnkiConnectWithURL returned nil")
	}
	if client.URL != customURL {
		t.Errorf("Expected URL %s, got %s", customURL, client.URL)
	}
}
