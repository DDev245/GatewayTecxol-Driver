package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/tecxol/gatewaytecxol-driver/internal/config"
	"go.uber.org/zap"
)

type Client struct {
	cfg    config.MQTT
	client mqtt.Client
	log    *zap.Logger
}

func New(cfg config.MQTT, log *zap.Logger) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetCleanSession(false).
		SetAutoReconnect(true).
		SetMaxReconnectDelay(5 * time.Minute).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetKeepAlive(time.Duration(cfg.Keepalive) * time.Second).
		SetOrderMatters(false)

	if cfg.TLS.Enabled {
		tlsCfg, err := buildTLSConfig(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("tls config: %w", err)
		}
		opts.SetTLSConfig(tlsCfg)
		if cfg.TLS.InsecureSkipVerify {
			log.Warn("mqtt TLS verification disabled (dev only)", zap.String("broker", cfg.Broker))
		}
	}

	c := &Client{cfg: cfg, log: log, client: mqtt.NewClient(opts)}
	c.client.Connect()
	return c, nil
}

func buildTLSConfig(cfg config.TLS) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	if cfg.CAFile != "" {
		caBytes, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("invalid ca pem")
		}
		tlsCfg.RootCAs = pool
	}

	if cfg.ClientCert != "" && cfg.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCert, cfg.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("load client keypair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}

func (c *Client) IsConnected() bool {
	return c.client != nil && c.client.IsConnected()
}

func (c *Client) Publish(topic string, payload []byte, retained bool) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("mqtt not connected")
	}
	tok := c.client.Publish(c.cfg.TopicPrefix+"/"+topic, c.cfg.QoS, retained, payload)
	if !tok.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("publish timeout")
	}
	return tok.Error()
}

func (c *Client) Disconnect(quiesce uint) {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(quiesce)
	}
}
