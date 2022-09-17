package pkg

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/sirupsen/logrus"
)

type MqttPublisher struct {
	// Core components
	cfg    *MqttConfig
	client mqtt.Client
}

func NewMqttPublisher(cfg *MqttConfig) (*MqttPublisher, error) {
	p := &MqttPublisher{
		cfg: cfg,
	}

	err := p.Setup()

	return p, err
}

func (p *MqttPublisher) Setup() error {
	opts := mqtt.NewClientOptions()
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(2 * time.Second)
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", p.cfg.Broker, p.cfg.Port))
	opts.SetClientID(fmt.Sprintf("gmc_mqtt_%s", p.cfg.Topic))
	if p.cfg.Username != "" {
		opts.SetUsername(p.cfg.Username)
	}
	if p.cfg.Password != "" {
		opts.SetPassword(p.cfg.Password)
	}
	opts.SetDefaultPublishHandler(p.onMessageReceived)
	opts.OnConnect = p.onConnectHandler
	opts.OnConnectionLost = p.onConnectionLostHandler
	p.client = mqtt.NewClient(opts)

	if token := p.client.Connect(); token.Wait() && token.Error() != nil {
		logrus.Errorf("[MQTT] Setup of mqtt publisher failed: %v!", token.Error())
		panic(token.Error())
	}

	logrus.Infof("[MQTT] Setup of mqtt publisher completed!")
	return nil
}

func (p *MqttPublisher) Close() {
	p.client.Disconnect(250)
}

func (p *MqttPublisher) onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	logrus.Infof("[MQTT] TOPIC: %s", msg.Topic())
	logrus.Infof("[MQTT] MSG: %s", msg.Payload())
}

func (p *MqttPublisher) onConnectHandler(_ mqtt.Client) {
	logrus.Infof("[MQTT] Connected to broker!")
}

func (p *MqttPublisher) onConnectionLostHandler(_ mqtt.Client, err error) {
	logrus.Warnf("[MQTT] Connection to broker lost: %v!", err)
}

func (p *MqttPublisher) Publish(temperature float64, cpm int, version string, isOnline bool) error {
	if err := p.publishHomeAssistantConfig(version); err != nil {
		return fmt.Errorf("failed to publish mqtt config: %w", err)
	}

	time.Sleep(2 * time.Second) // wait for home assistant to process new topics

	if err := p.publishTopics(temperature, cpm, isOnline); err != nil {
		return fmt.Errorf("failed to publish mqtt sensors: %w", err)
	}

	return nil
}

func (p *MqttPublisher) publishHomeAssistantConfig(version string) error {
	topicStatus := fmt.Sprintf("homeassistant/binary_sensor/%s/status/config", p.cfg.Topic)
	availabilityConfig := map[string]any{
		"name":               "Status",
		"state_topic":        fmt.Sprintf("gmc/%s/status", p.cfg.Topic),
		"availability_topic": fmt.Sprintf("gmc/%s/status", p.cfg.Topic),
		"device_class":       "connectivity",
		"payload_on":         "online",
		"payload_off":        "offline",
		"expire_after":       "240",
		"unique_id":          fmt.Sprintf("gmc_%s_status", p.cfg.Topic),
		"device": map[string]any{
			"identifiers":  p.cfg.Topic,
			"name":         p.cfg.Topic,
			"manufacturer": "GMC",
			"model":        "GMC-320 Plus",
			"sw_version":   version,
		},
	}

	payload, _ := json.Marshal(availabilityConfig)
	token := p.client.Publish(topicStatus, 0, false, string(payload))
	token.Wait()

	topicTemperature := fmt.Sprintf("homeassistant/sensor/%s/temperature/config", p.cfg.Topic)
	temperatureConfig := map[string]any{
		"name":                "Temperature",
		"state_topic":         fmt.Sprintf("gmc/%s/temperature", p.cfg.Topic),
		"availability_topic":  fmt.Sprintf("gmc/%s/status", p.cfg.Topic),
		"unit_of_measurement": "°C",
		"device_class":        "temperature",
		"state_class":         "measurement",
		"value_template":      "{{ value_json.value | float }}",
		"unique_id":           fmt.Sprintf("gmc_%s_temp", p.cfg.Topic),
		"device": map[string]any{
			"identifiers":  p.cfg.Topic,
			"name":         p.cfg.Topic,
			"manufacturer": "GMC",
			"model":        "GMC-320 Plus",
			"sw_version":   version,
		},
	}
	payload, _ = json.Marshal(temperatureConfig)
	token = p.client.Publish(topicTemperature, 0, false, string(payload))
	token.Wait()

	topicCpm := fmt.Sprintf("homeassistant/sensor/%s/cpm/config", p.cfg.Topic)
	cpmConfig := map[string]any{
		"name":                "CPM",
		"state_topic":         fmt.Sprintf("gmc/%s/cpm", p.cfg.Topic),
		"availability_topic":  fmt.Sprintf("gmc/%s/status", p.cfg.Topic),
		"unit_of_measurement": "CPM",
		"state_class":         "measurement",
		"value_template":      "{{ value_json.value | int }}",
		"unique_id":           fmt.Sprintf("gmc_%s_cpm", p.cfg.Topic),
		"device": map[string]any{
			"identifiers":  p.cfg.Topic,
			"name":         p.cfg.Topic,
			"manufacturer": "GMC",
			"model":        "GMC-320 Plus",
			"sw_version":   version,
		},
	}
	payload, _ = json.Marshal(cpmConfig)
	token = p.client.Publish(topicCpm, 0, false, string(payload))
	token.Wait()

	return nil
}

func (p *MqttPublisher) publishTopics(temperature float64, cpm int, isOnline bool) error {
	topicStatus := fmt.Sprintf("gmc/%s/status", p.cfg.Topic)
	status := "offline"
	if isOnline {
		status = "online"
	}
	token := p.client.Publish(topicStatus, 0, false, status)
	token.Wait()

	topicTemperature := fmt.Sprintf("gmc/%s/temperature", p.cfg.Topic)
	temperatureValue := map[string]any{
		"value": temperature,
		"unit":  "°C",
	}
	payload, _ := json.Marshal(temperatureValue)
	token = p.client.Publish(topicTemperature, 0, false, string(payload))
	token.Wait()

	topicCpm := fmt.Sprintf("gmc/%s/cpm", p.cfg.Topic)
	cpmValue := map[string]any{
		"value": cpm,
		"unit":  "CPM",
	}
	payload, _ = json.Marshal(cpmValue)
	token = p.client.Publish(topicCpm, 0, false, string(payload))
	token.Wait()
	return nil
}
