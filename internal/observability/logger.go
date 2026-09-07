package observability

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level, format string) (*zap.Logger, error) {
	lvl := zap.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	if format == "console" {
		cfg.Encoding = "console"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg.Encoding = "json"
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.MessageKey = "msg"
		cfg.EncoderConfig.LevelKey = "level"
	}

	return cfg.Build()
}
