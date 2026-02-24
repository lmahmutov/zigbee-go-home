//go:build !no_mqtt

package main

import (
	"log/slog"

	mqttbridge "zigbee-go-home/internal/mqtt"

	"zigbee-go-home/internal/coordinator"
)

type mqttStopper struct {
	bridge *mqttbridge.Bridge
}

func (m *mqttStopper) Stop() {
	if m.bridge != nil {
		m.bridge.Stop()
	}
}

func initMQTT(coord *coordinator.Coordinator, cfg *Config, logger *slog.Logger) *mqttStopper {
	if !cfg.MQTT.Enabled {
		return &mqttStopper{}
	}
	bridge, err := mqttbridge.NewBridge(coord, mqttbridge.Config{
		Broker:      cfg.MQTT.Broker,
		Username:    cfg.MQTT.Username,
		Password:    cfg.MQTT.Password,
		TopicPrefix: cfg.MQTT.TopicPrefix,
	}, logger)
	if err != nil {
		logger.Error("mqtt bridge", "err", err)
		return &mqttStopper{}
	}
	bridge.Start()
	return &mqttStopper{bridge: bridge}
}
