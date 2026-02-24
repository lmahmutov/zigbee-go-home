//go:build no_mqtt

package main

import (
	"log/slog"

	"zigbee-go-home/internal/coordinator"
)

type mqttStopper struct{}

func (m *mqttStopper) Stop() {}

func initMQTT(_ *coordinator.Coordinator, _ *Config, _ *slog.Logger) *mqttStopper {
	return &mqttStopper{}
}
