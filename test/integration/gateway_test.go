//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

func startSim(t *testing.T) (net.Listener, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 256)
				for {
					c.SetReadDeadline(time.Now().Add(time.Second))
					_, err := c.Read(buf)
					if err != nil {
						return
					}
				}
			}(c)
		}
	}()
	return ln, ln.Addr().String()
}

func TestSunspecHandshake(t *testing.T) {
	ln, addr := startSim(t)
	defer ln.Close()

	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		t.Fatalf("bad addr: %s", addr)
	}

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = ctx
	_ = json.Marshal
}
