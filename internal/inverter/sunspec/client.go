package sunspec

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/tecxol/gatewaytecxol-driver/internal/config"
	"github.com/tecxol/gatewaytecxol-driver/internal/inverter"
	"go.uber.org/zap"
)

const (
	modbusFuncReadHolding = 0x03
	sunspecBaseRegister   = 40000
	sunspecMagic          = 0x53756e53
)

type Client struct {
	addr   string
	unitID byte
	conn   net.Conn
	mu     sync.Mutex
	models map[int]bool
	log    *zap.Logger
	host   string
	port   int
	device string
}

func New(cfg config.SunSpec, log *zap.Logger) (inverter.Driver, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("sunspec.host is required")
	}
	if cfg.Port == 0 {
		return nil, fmt.Errorf("sunspec.port is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 2 * time.Second
	}
	models := map[int]bool{}
	if len(cfg.Models) == 0 {
		models[101] = true
	} else {
		for _, m := range cfg.Models {
			models[m] = true
		}
	}
	return &Client{
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		host:   cfg.Host,
		port:   cfg.Port,
		unitID: cfg.UnitID,
		models: models,
		log:    log,
		device: fmt.Sprintf("sunspec-%s", cfg.Host),
	}, nil
}

func (c *Client) Model() string   { return "sunspec" }
func (c *Client) DeviceID() string { return c.device }
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) connect(ctx context.Context) error {
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Scan(ctx context.Context) (inverter.Snapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		if err := c.connect(ctx); err != nil {
			return inverter.Snapshot{}, err
		}
	}

	if err := c.conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return inverter.Snapshot{}, err
	}

	base, err := c.readBase(ctx)
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return inverter.Snapshot{}, fmt.Errorf("read base: %w", err)
	}

	metrics := map[string]interface{}{}
	for modelID := range c.models {
		values, err := c.readModel(modelID)
		if err != nil {
			c.log.Warn("read model failed", zap.Int("model", modelID), zap.Error(err))
			continue
		}
		for k, v := range values {
			metrics[k] = v
		}
	}

	return inverter.Snapshot{
		Timestamp: time.Now().UTC(),
		Model:     c.Model(),
		DeviceID:  c.device,
		Metrics:   metrics,
		Status:    "ok",
	}, nil
}

func (c *Client) readBase(ctx context.Context) (uint16, error) {
	req := c.buildRequest(sunspecBaseRegister, 2)
	if _, err := c.conn.Write(req); err != nil {
		return 0, err
	}
	hdr := make([]byte, 8)
	if _, err := readFull(c.conn, hdr); err != nil {
		return 0, err
	}
	magic := binary.BigEndian.Uint32(hdr[0:4])
	if magic != sunspecMagic {
		return 0, fmt.Errorf("invalid sunSpec magic: 0x%x", magic)
	}
	return binary.BigEndian.Uint16(hdr[6:8]), nil
}

func (c *Client) readModel(id int) (map[string]interface{}, error) {
	values := map[string]interface{}{}
	switch id {
	case 101:
		values["ac.w"] = 0.0
		values["ac.var"] = 0.0
		values["ac.va"] = 0.0
		values["freq"] = 60.0
	case 103:
		values["status"] = uint16(1)
	case 120, 121, 123:
		values["dc.w"] = 0.0
		values["dc.v"] = 0.0
		values["dc.a"] = 0.0
	}
	return values, nil
}

func (c *Client) buildRequest(reg uint16, qty uint16) []byte {
	req := make([]byte, 8)
	binary.BigEndian.PutUint16(req[0:2], reg)
	binary.BigEndian.PutUint16(req[2:4], modbusFuncReadHolding)
	binary.BigEndian.PutUint16(req[4:6], qty)
	binary.BigEndian.PutUint16(req[6:8], 0)
	return c.appendCRC(req)
}

func (c *Client) appendCRC(pdu []byte) []byte {
	crc := modbusCRC(pdu)
	out := make([]byte, len(pdu)+2)
	copy(out, pdu)
	binary.LittleEndian.PutUint16(out[len(pdu):], crc)
	return out
}

func modbusCRC(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func readFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
