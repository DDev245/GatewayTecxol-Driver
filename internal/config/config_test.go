package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := `mqtt:
  broker: "tcp://localhost:1883"
  client_id: "test"
  topic_prefix: "test/prefix"
inverter:
  type: "sunspec"
  sunspec:
    host: "127.0.0.1"
    port: 502
`
	path := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if c.MQTT.TopicPrefix != "test/prefix" {
		t.Fatalf("prefix: %s", c.MQTT.TopicPrefix)
	}
}

func TestLoadMissingBroker(t *testing.T) {
	dir := t.TempDir()
	cfg := `mqtt:
  topic_prefix: "x"
inverter:
  type: "sunspec"
`
	path := filepath.Join(dir, "c.yaml")
	_ = os.WriteFile(path, []byte(cfg), 0o600)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}
