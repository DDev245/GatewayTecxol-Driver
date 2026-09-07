package inverter

import (
	"fmt"

	"github.com/tecxol/gatewaytecxol-driver/internal/config"
	"github.com/tecxol/gatewaytecxol-driver/internal/inverter/sunspec"
	"go.uber.org/zap"
)

func New(cfg config.Inverter, log *zap.Logger) (Driver, error) {
	switch cfg.Type {
	case "sunspec":
		return sunspec.New(cfg.SunSpec, log)
	default:
		return nil, fmt.Errorf("unknown inverter type: %s", cfg.Type)
	}
}
