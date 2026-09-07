package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Version  string   `mapstructure:"version"`
	Log      Log      `mapstructure:"log"`
	HTTP     HTTP     `mapstructure:"http"`
	MQTT     MQTT     `mapstructure:"mqtt"`
	Inverter Inverter `mapstructure:"inverter"`
}

type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type HTTP struct {
	Addr string `mapstructure:"addr"`
}

type MQTT struct {
	Broker      string  `mapstructure:"broker"`
	ClientID    string  `mapstructure:"client_id"`
	TopicPrefix string  `mapstructure:"topic_prefix"`
	QoS         byte    `mapstructure:"qos"`
	Keepalive   int     `mapstructure:"keepalive"`
	TLS         TLS     `mapstructure:"tls"`
}

type TLS struct {
	Enabled            bool   `mapstructure:"enabled"`
	CAFile             string `mapstructure:"ca_file"`
	ClientCert         string `mapstructure:"client_cert"`
	ClientKey          string `mapstructure:"client_key"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

type Inverter struct {
	Type    string    `mapstructure:"type"`
	SunSpec SunSpec   `mapstructure:"sunspec"`
	Poll    Poll      `mapstructure:"poll"`
}

type SunSpec struct {
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	UnitID  byte          `mapstructure:"unit_id"`
	Timeout time.Duration `mapstructure:"timeout"`
	Models  []int         `mapstructure:"models"`
}

type Poll struct {
	Interval time.Duration `mapstructure:"interval"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("GATEWAY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.MQTT.Broker == "" {
		return fmt.Errorf("mqtt.broker is required")
	}
	if c.MQTT.TopicPrefix == "" {
		return fmt.Errorf("mqtt.topic_prefix is required")
	}
	if c.Inverter.Type == "" {
		return fmt.Errorf("inverter.type is required")
	}
	if c.HTTP.Addr == "" {
		c.HTTP.Addr = ":9091"
	}
	return nil
}
