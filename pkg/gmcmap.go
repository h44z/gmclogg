package pkg

import (
	"fmt"
	"net/http"
	"time"
)

type GmcMap struct {
	cfg *GmcMapConfig

	client *http.Client
}

func NewGmcMap(cfg *GmcMapConfig) *GmcMap {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &GmcMap{
		cfg:    cfg,
		client: client,
	}
}

func (g *GmcMap) Publish(temperature float64, cpm int, version string, isOnline bool) error {
	if !isOnline {
		return nil // nothing to publish
	}

	url := fmt.Sprintf("%s?AID=%s&GID=%s&CPM=%d", g.cfg.BaseUrl, g.cfg.UserId, g.cfg.GeigerCounterId, cpm)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("gmcmap request invalid: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gmcmap request failed: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("gmcmap request failed with status: %d", res.StatusCode)
	}

	return nil
}
