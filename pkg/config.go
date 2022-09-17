package pkg

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type GmcConfig struct {
	SerialPort  string `envconfig:"SERIAL_PORT"`
	SerialBaud  int    `envconfig:"SERIAL_BAUD"`
	PollingRate int    `envconfig:"POLLING_RATE"` // Seconds
}

func NewGmcConfig() *GmcConfig {
	// Default config
	cfg := &GmcConfig{
		SerialPort:  "/dev/ttyUSB_RAD",
		SerialBaud:  115200,
		PollingRate: 60, // 1 Minute
	}
	if err := loadConfigEnv(cfg); err != nil {
		logrus.Warnf("unable to load environment config: %v", err)
	}

	return cfg
}

type InfluxConfig struct {
	URL      string `envconfig:"INFLUX_URL"`
	UserName string `envconfig:"INFLUX_USER"`
	Password string `envconfig:"INFLUX_PASS"`
	Bucket   string `envconfig:"INFLUX_BUCKET"`
}

func NewInfluxConfig() *InfluxConfig {
	// Default config
	cfg := &InfluxConfig{
		URL:      "http://localhost:8086",
		UserName: "influxuser",
		Password: "influxpass",
		Bucket:   "gmclogg",
	}
	if err := loadConfigEnv(cfg); err != nil {
		logrus.Warnf("unable to load environment config: %v", err)
	}

	return cfg
}

type MqttConfig struct {
	Broker   string `envconfig:"MQTT_BROKER"`
	Port     int    `envconfig:"MQTT_PORT"`
	Username string `envconfig:"MQTT_USER"`
	Password string `envconfig:"MQTT_PASS"`

	Topic string `envconfig:"MQTT_TOPIC"`
}

func NewMqttConfig() *MqttConfig {
	// Default config
	cfg := &MqttConfig{
		Broker:   "localhost",
		Port:     1883,
		Username: "mqttUser",
		Password: "mqttPassword",
		Topic:    "gmclogg",
	}
	if err := loadConfigEnv(cfg); err != nil {
		logrus.Warnf("unable to load environment config: %v", err)
	}

	return cfg
}

type GmcMapConfig struct {
	BaseUrl         string `envconfig:"GMC_MAP_URL"`
	UserId          string `envconfig:"GMC_MAP_USER"`
	GeigerCounterId string `envconfig:"GMC_MAP_GEIGER_COUNTER"`
}

func NewGmcMapConfig() *GmcMapConfig {
	// Default config
	cfg := &GmcMapConfig{
		BaseUrl:         "http://www.GMCmap.com/log2.asp",
		UserId:          "123456",
		GeigerCounterId: "789456123",
	}
	if err := loadConfigEnv(cfg); err != nil {
		logrus.Warnf("unable to load environment config: %v", err)
	}

	return cfg
}

func loadConfigEnv(cfg any) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return err
	}

	return nil
}
