//go:build !no_automation

package main

import (
	"log/slog"
	"time"

	"zigbee-go-home/internal/automation"
	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/web"
)

type autoStopper struct {
	engine *automation.Engine
}

func (a *autoStopper) Stop() {
	if a.engine != nil {
		a.engine.Stop()
	}
}

func initAutomation(coord *coordinator.Coordinator, cfg *Config, logger *slog.Logger) (*autoStopper, []web.ServerOption) {
	scriptMgr, err := automation.NewManager(cfg.ScriptsDir)
	if err != nil {
		logger.Error("create script manager", "err", err)
		return &autoStopper{}, nil
	}

	execTimeout := 10 * time.Second
	if cfg.Exec.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Exec.Timeout); err == nil {
			execTimeout = d
		} else {
			logger.Warn("invalid exec.timeout, using default", "value", cfg.Exec.Timeout, "default", execTimeout)
		}
	}

	engine := automation.NewEngine(coord, scriptMgr, logger,
		automation.SystemConfig{
			ExecAllowlist: cfg.Exec.Allowlist,
			ExecTimeout:   execTimeout,
		},
		automation.TelegramConfig{
			BotToken: cfg.Telegram.BotToken,
			ChatIDs:  cfg.Telegram.ChatIDs,
		},
	)
	engine.Start()

	opts := []web.ServerOption{
		web.WithAutomation(engine, scriptMgr),
	}
	return &autoStopper{engine: engine}, opts
}
