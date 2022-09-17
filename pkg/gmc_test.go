//go:build integration

package pkg

import (
	"testing"
)

func TestGmc_FetchCpm(t *testing.T) {
	g := NewGmc(&GmcConfig{
		SerialPort: "/dev/ttyUSB1",
		SerialBaud: 115200,
	})

	t.Log("Opening COM port")
	err := g.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	t.Log("Starting to fetch...")

	version, err := g.FetchVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Version:", version)

	cpm, err := g.FetchCpm()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("CPM:", cpm)

	temperature, err := g.FetchTemperature()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Temperature:", temperature)
}
