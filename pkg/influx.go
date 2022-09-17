package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxLogger struct {
	cfg    *InfluxConfig
	client influxdb2.Client
}

func NewInfluxLogger(cfg *InfluxConfig) *InfluxLogger {
	i := &InfluxLogger{
		cfg: cfg,
	}
	i.client = influxdb2.NewClient(cfg.URL, fmt.Sprintf("%s:%s", cfg.UserName, cfg.Password))
	return i
}

func (l *InfluxLogger) Close() {
	if l.client != nil {
		l.client.Close()
	}
}

func (l *InfluxLogger) logPoints(bucket string, points ...*write.Point) error {
	writeAPI := l.client.WriteAPIBlocking("", bucket)

	// Write data
	err := writeAPI.WritePoint(context.Background(), points...)
	if err != nil {
		return fmt.Errorf("failed to write influx points: %w", err)
	}

	return nil
}

func (l *InfluxLogger) Publish(temperature float64, cpm int, version string, isOnline bool) error {
	if !isOnline {
		return nil // nothing to publish
	}

	points := make([]*write.Point, 0, 2)
	points = append(points, influxdb2.NewPoint("temperature", // Measurement
		map[string]string{"unit": "°C", "location": "Vill"}, // Tags
		map[string]any{"value": temperature},                // Fields
		time.Now()))
	points = append(points, influxdb2.NewPoint("cpm", // Measurement
		map[string]string{"unit": "CPM", "location": "Vill"}, // Tags
		map[string]any{"value": cpm},                         // Fields
		time.Now()))

	if err := l.logPoints(l.cfg.Bucket, points...); err != nil {
		return err
	}

	return nil
}
