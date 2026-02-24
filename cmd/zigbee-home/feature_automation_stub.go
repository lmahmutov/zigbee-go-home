//go:build no_automation

package main

import (
	"log/slog"

	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/web"
)

type autoStopper struct{}

func (a *autoStopper) Stop() {}

func initAutomation(_ *coordinator.Coordinator, _ *Config, _ *slog.Logger) (*autoStopper, []web.ServerOption) {
	return &autoStopper{}, nil
}
