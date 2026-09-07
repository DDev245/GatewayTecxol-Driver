package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tecxol/gatewaytecxol-driver/internal/config"
	"github.com/tecxol/gatewaytecxol-driver/internal/health"
	"github.com/tecxol/gatewaytecxol-driver/internal/inverter"
	"github.com/tecxol/gatewaytecxol-driver/internal/mqtt"
	"github.com/tecxol/gatewaytecxol-driver/internal/observability"
	"go.uber.org/zap"
)

var version = "dev"

const schemaVersion = "0.1.0"

func run(logger *zap.Logger, cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	logger.Info("config loaded",
		zap.String("mqtt", cfg.MQTT.Broker),
		zap.String("inverter", cfg.Inverter.Type),
		zap.String("prefix", cfg.MQTT.TopicPrefix),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mq, err := mqtt.New(cfg.MQTT, logger)
	if err != nil {
		return fmt.Errorf("mqtt: %w", err)
	}
	defer mq.Disconnect(2000)

	drv, err := inverter.New(cfg.Inverter, logger)
	if err != nil {
		return fmt.Errorf("inverter: %w", err)
	}
	defer drv.Close()

	status := health.New()

	go pollLoop(ctx, drv, mq, status, logger, cfg.Inverter.Poll.Interval)

	srv := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: health.NewServer(cfg.HTTP.Addr, status).Routes(),
	}

	go func() {
		logger.Info("http server starting", zap.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server", zap.Error(err))
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	cancel()
	return nil
}

func pollLoop(ctx context.Context, drv inverter.Driver, mq *mqtt.Client, status *health.Status, log *zap.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scanCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			snap, err := drv.Scan(scanCtx)
			cancel()
			status.SetBroker(mq.IsConnected())
			health.SetBrokerUp(mq.IsConnected())
			status.RecordPoll(err)
			health.IncPolls()
			if err != nil {
				health.IncErrors()
				log.Warn("poll error", zap.Error(err))
				continue
			}
			health.SetLastPoll(snap.Timestamp)

			payload, err := encode(snap)
			if err != nil {
				log.Warn("encode payload", zap.Error(err))
				continue
			}
			if err := mq.Publish("telemetry", payload, false); err != nil {
				health.IncErrors()
				log.Warn("publish telemetry", zap.Error(err))
				continue
			}
			log.Debug("telemetry published", zap.Int("bytes", len(payload)))
		}
	}
}

func encode(snap inverter.Snapshot) ([]byte, error) {
	wrapped := struct {
		SchemaVersion string             `json:"schema_version"`
		Snapshot      inverter.Snapshot  `json:"snapshot"`
	}{
		SchemaVersion: schemaVersion,
		Snapshot:      snap,
	}
	return json.Marshal(wrapped)
}
