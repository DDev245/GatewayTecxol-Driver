package inverter

import (
	"context"
	"time"
)

type Snapshot struct {
	Timestamp time.Time              `json:"ts"`
	Model     string                 `json:"model"`
	DeviceID  string                 `json:"dev"`
	Metrics   map[string]interface{} `json:"metrics"`
	Status    string                 `json:"status"`
}

type Driver interface {
	Model() string
	DeviceID() string
	Scan(ctx context.Context) (Snapshot, error)
	Close() error
}

type Factory func(ctx context.Context) (Driver, error)
