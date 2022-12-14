package main

import (
	"os"
	"strconv"
	"time"

	"github.com/h44z/gmclogg/pkg"
	"github.com/sirupsen/logrus"
)

type publisher interface {
	Publish(temperature float64, cpm int, version string, isOnline bool) error
}

// You need to create a influx db before using this tool:
/*
$ influx (-username xxx -password yyy)
> create database gmclogg
> exit
*/
func main() {
	logrus.SetLevel(logrus.DebugLevel)

	gCfg := pkg.NewGmcConfig()

	gmc := pkg.NewGmc(gCfg)
	if err := gmc.Open(); err != nil {
		logrus.Fatal("[MAIN] Unable to initialize GMC!", err)
	}
	defer gmc.Close()

	gmcmap, mqtt, influx := features()

	var publishers []publisher

	if gmcmap {
		mCfg := pkg.NewGmcMapConfig()
		m := pkg.NewGmcMap(mCfg)

		publishers = append(publishers, m)
	}

	if mqtt {
		mCfg := pkg.NewMqttConfig()
		p, err := pkg.NewMqttPublisher(mCfg)
		if err != nil {
			logrus.Fatalf("[MAIN] Unable to initialize MQTT publisher: %v", err)
		}
		defer p.Close()

		publishers = append(publishers, p)
	}

	if influx {
		iCfg := pkg.NewInfluxConfig()
		i := pkg.NewInfluxLogger(iCfg)
		defer i.Close()

		publishers = append(publishers, i)
	}

	logrus.Infof("[MAIN] Starting in %v (%d pub)...", time.Duration(gCfg.PollingRate)*time.Second, len(publishers))

	// Start ticker
	ticker := time.NewTicker(time.Duration(gCfg.PollingRate) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logrus.Info("[MAIN] Tick started...")
			err := gmc.FlushBus()
			if err != nil {
				logrus.Errorf("[MAIN] GMC error: %v", err)
			}

			isOnline := true
			version, err := gmc.FetchVersion()
			if err != nil {
				logrus.Errorf("[MAIN] Lost connection to GMC: %v", err)
				_ = gmc.Reconnect()
				isOnline = false
			}
			logrus.Debugf("[MAIN] fetched version %s", version)
			cpm, err := gmc.FetchCpm()
			if err != nil {
				logrus.Errorf("[MAIN] Lost connection to GMC: %v", err)
				_ = gmc.Reconnect()
				isOnline = false
			}
			logrus.Debugf("[MAIN] fetched cpm %d", cpm)
			temperature, err := gmc.FetchTemperature()
			if err != nil {
				logrus.Errorf("[MAIN] Lost connection to GMC: %v", err)
				_ = gmc.Reconnect()
				isOnline = false
			}
			logrus.Debugf("[MAIN] fetched temperature %f", temperature)

			logrus.Infof("[MAIN] Fetched: %0.2f ??C, %d CPM, %s (online: %t)", temperature, cpm, version, isOnline)

			for i, p := range publishers {
				err := p.Publish(temperature, cpm, version, isOnline)
				if err != nil {
					logrus.Errorf("[MAIN] Failed to publish: %v", err)
				}
				logrus.Debugf("[MAIN] Published #%d", i)
			}

			logrus.Info("[MAIN] Tick completed!")
		}
	}
}

func features() (gmcmap, mqtt, influx bool) {
	gmcmap = true
	mqtt = true
	influx = true

	if val, err := strconv.ParseBool(os.Getenv("ENABLE_GMC_MAP")); err == nil && !val {
		gmcmap = false
	}
	if val, err := strconv.ParseBool(os.Getenv("ENABLE_MQTT")); err == nil && !val {
		mqtt = false
	}
	if val, err := strconv.ParseBool(os.Getenv("ENABLE_INFLUX")); err == nil && !val {
		influx = false
	}
	return
}
