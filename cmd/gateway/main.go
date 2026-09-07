package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tecxol/gatewaytecxol-driver/internal/config"
	"github.com/tecxol/gatewaytecxol-driver/internal/observability"
	"go.uber.org/zap"
)

func newRootCmd() *cobra.Command {
	var cfgFile string
	var logLevel string
	var logFormat string

	root := &cobra.Command{
		Use:           "gateway",
		Short:         "Tecxol gateway driver (SunSpec -> MQTT 5)",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "config file")
	root.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level")
	root.PersistentFlags().StringVar(&logFormat, "log-format", "json", "log format (json|console)")

	logger := func() *zap.Logger {
		l, _ := observability.NewLogger(logLevel, logFormat)
		return l
	}

	root.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print version",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("gateway %s\n", version)
			},
		},
		&cobra.Command{
			Use:   "run",
			Short: "Run the gateway",
			RunE: func(cmd *cobra.Command, args []string) error {
				return run(logger(), cfgFile)
			},
		},
		&cobra.Command{
			Use:   "doctor",
			Short: "Run diagnostics on config and broker reachability",
			RunE: func(cmd *cobra.Command, args []string) error {
				return doctor(cfgFile, logger())
			},
		},
	)

	_ = viper.BindPFlags(root.PersistentFlags())
	return root
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func doctor(cfgPath string, log *zap.Logger) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	failed := 0

	if err := checkBroker(cfg.MQTT.Broker, 3*time.Second); err != nil {
		log.Error("broker unreachable", zap.String("broker", cfg.MQTT.Broker), zap.Error(err))
		failed++
	} else {
		log.Info("broker reachable", zap.String("broker", cfg.MQTT.Broker))
	}

	d := net.JoinHostPort(cfg.Inverter.SunSpec.Host, fmt.Sprintf("%d", cfg.Inverter.SunSpec.Port))
	if err := checkInverter(d, 2*time.Second); err != nil {
		log.Warn("inverter unreachable", zap.String("addr", d), zap.Error(err))
	} else {
		log.Info("inverter reachable", zap.String("addr", d))
	}

	if cfg.MQTT.TLS.Enabled {
		for _, p := range []string{cfg.MQTT.TLS.CAFile, cfg.MQTT.TLS.ClientCert, cfg.MQTT.TLS.ClientKey} {
			if p == "" {
				continue
			}
			if _, err := os.Stat(p); err != nil {
				log.Error("tls file missing", zap.String("path", p))
				failed++
			}
		}
	}

	if failed > 0 {
		return fmt.Errorf("doctor: %d check(s) failed", failed)
	}
	return nil
}

func checkBroker(broker string, timeout time.Duration) error {
	u, err := url.Parse(broker)
	if err != nil {
		return err
	}
	host := u.Host
	if u.Scheme == "tls" || u.Scheme == "ssl" {
		if _, _, err := net.SplitHostPort(host); err != nil {
			host = host + ":8883"
		}
	}
	if u.Scheme == "tcp" {
		if _, _, err := net.SplitHostPort(host); err != nil {
			host = host + ":1883"
		}
	}
	c, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return err
	}
	return c.Close()
}

func checkInverter(addr string, timeout time.Duration) error {
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	return c.Close()
}
