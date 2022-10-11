package pkg

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/tarm/serial"
)

// Gmc implements the protocol from here: http://www.gqelectronicsllc.com/download/GQ-RFC1201.txt
type Gmc struct {
	cfg  *GmcConfig
	conn *serial.Port

	mux sync.Mutex
}

func NewGmc(cfg *GmcConfig) *Gmc {
	r := &Gmc{
		cfg: cfg,
		mux: sync.Mutex{},
	}

	return r
}

func (g *Gmc) Open() error {
	c := &serial.Config{Name: g.cfg.SerialPort, Baud: g.cfg.SerialBaud, ReadTimeout: time.Millisecond * 1000}
	var err error
	g.conn, err = serial.OpenPort(c)
	if err != nil {
		return err
	}

	err = g.DisableHeartBeat() // disable heartbeat, we use polling
	if err != nil {
		return fmt.Errorf("heartbeat error: %w", err)
	}

	return nil
}

func (g *Gmc) Close() {
	if g.conn != nil {
		_ = g.conn.Close()
	}
}

func (g *Gmc) Reconnect() error {
	g.Close()
	err := g.Open()
	if err != nil {
		return err
	}

	return nil
}

func (g *Gmc) FetchCpm() (int, error) {
	g.mux.Lock()
	defer g.mux.Unlock()

	_, err := g.conn.Write([]byte("<GETCPM>>"))
	if err != nil {
		return 0, fmt.Errorf("failed to send GETCPM command: %w", err)
	}

	buf := make([]byte, 2)
	_, err = g.conn.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to read GETCPM response: %w", err)
	}

	cpm := int(buf[0])<<8 | int(buf[1]) // buf[0] * 256 + buf[1];

	return cpm, nil
}

func (g *Gmc) FetchTemperature() (float64, error) {
	g.mux.Lock()
	defer g.mux.Unlock()

	_, err := g.conn.Write([]byte("<GETTEMP>>"))
	if err != nil {
		return 0, fmt.Errorf("failed to send GETTEMP command: %w", err)
	}

	buf := make([]byte, 4)
	_, err = g.conn.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to read GETTEMP response: %w", err)
	}

	sign := 1.0
	if buf[2] != 0 {
		sign = -1.0
	}

	temp, err := strconv.ParseFloat(fmt.Sprintf("%d.%d", buf[0], buf[1]), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse GETTEMP response: %w", err)
	}
	temp *= sign

	return temp, nil
}

func (g *Gmc) FetchVersion() (string, error) {
	g.mux.Lock()
	defer g.mux.Unlock()

	_, err := g.conn.Write([]byte("<GETVER>>"))
	if err != nil {
		return "", fmt.Errorf("failed to send GETVER command: %w", err)
	}

	buf := make([]byte, 14)
	_, err = g.conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read GETVER response: %w", err)
	}

	return string(buf), nil
}

func (g *Gmc) DisableHeartBeat() error {
	g.mux.Lock()
	defer g.mux.Unlock()

	_, err := g.conn.Write([]byte("<HEARTBEAT0>>"))
	if err != nil {
		return fmt.Errorf("failed to send HEARTBEAT0 command: %w", err)
	}

	err = g.conn.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush HEARTBEAT0 response: %w", err)
	}

	return nil
}

func (g *Gmc) FlushBus() error {
	g.mux.Lock()
	defer g.mux.Unlock()

	err := g.conn.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush bus: %w", err)
	}

	return nil
}
