package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/ncp"
	"zigbee-go-home/internal/store"
	"zigbee-go-home/internal/web"
	"zigbee-go-home/internal/zcl"
	"zigbee-go-home/internal/zcl/clusters"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

type Config struct {
	NCP struct {
		Type string `yaml:"type"` // "nrf52840"
		Port string `yaml:"port"`
		Baud int    `yaml:"baud"`
	} `yaml:"ncp"`
	Network struct {
		Channel  uint8  `yaml:"channel"`
		PanID    uint16 `yaml:"pan_id"`
		ExtPanID string `yaml:"extended_pan_id"`
	} `yaml:"network"`
	Web struct {
		Listen         string   `yaml:"listen"`
		APIKey         string   `yaml:"api_key"`
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"web"`
	Store struct {
		Path string `yaml:"path"`
	} `yaml:"store"`
	MQTT struct {
		Enabled     bool   `yaml:"enabled"`
		Broker      string `yaml:"broker"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		TopicPrefix string `yaml:"topic_prefix"`
	} `yaml:"mqtt"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"log"`
	Telegram struct {
		BotToken string   `yaml:"bot_token"`
		ChatIDs  []string `yaml:"chat_ids"`
	} `yaml:"telegram"`
	Exec struct {
		Allowlist []string `yaml:"allowlist"`
		Timeout   string   `yaml:"timeout"`
	} `yaml:"exec"`
	DevicesDir string `yaml:"devices_dir"`
	ScriptsDir string `yaml:"scripts_dir"`
}

func (c *Config) validate() error {
	if c.NCP.Port == "" {
		return fmt.Errorf("ncp.port is required")
	}
	if c.Network.Channel < 11 || c.Network.Channel > 26 {
		return fmt.Errorf("network.channel must be 11-26, got %d", c.Network.Channel)
	}
	if c.Network.PanID == 0 || c.Network.PanID == 0xFFFF {
		return fmt.Errorf("network.pan_id must not be 0x0000 or 0xFFFF")
	}
	return nil
}

func main() {
	// Temporary logger for config loading errors.
	bootLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		bootLogger.Error("load config", "err", err)
		os.Exit(1)
	}

	if err := cfg.validate(); err != nil {
		bootLogger.Error("invalid config", "err", err)
		os.Exit(1)
	}

	// Create configured logger.
	logger := newLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("zigbee-go-home starting", "version", version)

	// Initialize ZCL registry
	registry := zcl.NewRegistry(logger)
	registerStandardClusters(registry)

	// Load device definitions (custom clusters + device configs) from devices directory.
	deviceDB, err := coordinator.LoadDeviceDir(cfg.DevicesDir, registry, logger)
	if err != nil {
		logger.Error("load device definitions", "err", err)
		os.Exit(1)
	}
	logger.Info("ZCL registry initialized", "clusters", len(registry.All()), "devices", deviceDB.Len())

	// Open store
	db, err := store.NewBoltStore(cfg.Store.Path)
	if err != nil {
		logger.Error("open store", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create NCP backend based on config
	backend, err := createNCP(cfg, logger)
	if err != nil {
		logger.Error("create NCP backend", "err", err)
		os.Exit(1)
	}
	defer backend.Close()

	// Parse extended PAN ID
	extPanID, err := coordinator.ParseExtPanID(cfg.Network.ExtPanID)
	if err != nil {
		logger.Error("parse ext pan id", "err", err)
		os.Exit(1)
	}

	// Create coordinator
	events := coordinator.NewEventBus(logger)
	coord := coordinator.New(backend, db, registry, deviceDB, events, coordinator.Config{
		Channel:  cfg.Network.Channel,
		PanID:    cfg.Network.PanID,
		ExtPanID: extPanID,
	}, coordinator.NCPConfig{
		Type: cfg.NCP.Type,
		Port: cfg.NCP.Port,
		Baud: cfg.NCP.Baud,
	}, logger)

	// Start coordinator
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := coord.Start(ctx); err != nil {
		logger.Error("start coordinator", "err", err)
		cancel()
		backend.Close()
		os.Exit(1)
	}
	cancel()

	// Start automation engine (no-op when built with no_automation tag).
	auto, autoWebOpts := initAutomation(coord, cfg, logger)

	// Start web server
	var webOpts []web.ServerOption
	if cfg.DevicesDir != "" {
		webOpts = append(webOpts, web.WithDevicesDir(cfg.DevicesDir))
	}
	if cfg.Web.APIKey != "" {
		webOpts = append(webOpts, web.WithAPIKey(cfg.Web.APIKey))
	}
	if len(cfg.Web.AllowedOrigins) > 0 {
		webOpts = append(webOpts, web.WithAllowedOrigins(cfg.Web.AllowedOrigins))
	}
	webOpts = append(webOpts, web.WithVersion(version))
	webOpts = append(webOpts, autoWebOpts...)

	webServer, err := web.NewServer(coord, logger, webOpts...)
	if err != nil {
		logger.Error("create web server", "err", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:         cfg.Web.Listen,
		Handler:      webServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("web server starting", "addr", cfg.Web.Listen)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server", "err", err)
		}
	}()

	// Start MQTT bridge (no-op when built with no_mqtt tag).
	mqtt := initMQTT(coord, cfg, logger)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	signal.Stop(sigCh)
	logger.Info("shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	auto.Stop()
	mqtt.Stop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown", "err", err)
	}
	webServer.Stop()
	coord.Stop()

	logger.Info("goodbye")
}

func createNCP(cfg *Config, logger *slog.Logger) (ncp.NCP, error) {
	switch cfg.NCP.Type {
	case "nrf52840", "":
		logger.Info("using nRF52840 NCP (ZBOSS/HDLC)", "port", cfg.NCP.Port, "baud", cfg.NCP.Baud)
		return ncp.NewNRF52840NCP(cfg.NCP.Port, cfg.NCP.Baud, logger)
	default:
		return nil, fmt.Errorf("unknown NCP type: %q (supported: nrf52840)", cfg.NCP.Type)
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Web.Listen == "" {
		cfg.Web.Listen = "127.0.0.1:8080"
	}
	if cfg.Store.Path == "" {
		cfg.Store.Path = "zigbee-home.db"
	}
	if cfg.NCP.Baud == 0 {
		cfg.NCP.Baud = 460800
	}
	if cfg.DevicesDir == "" {
		cfg.DevicesDir = "devices"
	}
	if cfg.ScriptsDir == "" {
		cfg.ScriptsDir = "scripts"
	}
	if cfg.MQTT.TopicPrefix == "" {
		cfg.MQTT.TopicPrefix = "zigbee2mqtt"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
	return &cfg, nil
}

func newLogger(cfg *Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Log.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch strings.ToLower(cfg.Log.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func registerStandardClusters(r *zcl.Registry) {
	// General (0x0000–0x00FF)
	r.Register(clusters.Basic)                          // 0x0000
	r.Register(clusters.PowerConfiguration)              // 0x0001
	r.Register(clusters.DeviceTemperatureConfiguration)  // 0x0002
	r.Register(clusters.Identify)                        // 0x0003
	r.Register(clusters.Groups)                          // 0x0004
	r.Register(clusters.Scenes)                          // 0x0005
	r.Register(clusters.OnOff)                           // 0x0006
	r.Register(clusters.OnOffSwitchConfiguration)        // 0x0007
	r.Register(clusters.LevelControl)                    // 0x0008
	r.Register(clusters.Alarms)                          // 0x0009
	r.Register(clusters.Time)                            // 0x000A
	r.Register(clusters.RSSILocation)                    // 0x000B
	r.Register(clusters.AnalogInput)                     // 0x000C
	r.Register(clusters.AnalogOutput)                    // 0x000D
	r.Register(clusters.AnalogValue)                     // 0x000E
	r.Register(clusters.BinaryInput)                     // 0x000F
	r.Register(clusters.BinaryOutput)                    // 0x0010
	r.Register(clusters.BinaryValue)                     // 0x0011
	r.Register(clusters.MultistateInput)                 // 0x0012
	r.Register(clusters.MultistateOutput)                // 0x0013
	r.Register(clusters.MultistateValue)                 // 0x0014
	r.Register(clusters.Commissioning)                   // 0x0015
	r.Register(clusters.OTAUpgrade)                      // 0x0019
	r.Register(clusters.PowerProfile)                    // 0x001A
	r.Register(clusters.ApplianceControl)                // 0x001B
	r.Register(clusters.PollControl)                     // 0x0020
	r.Register(clusters.GreenPower)                      // 0x0021

	// Closures (0x0100–0x01FF)
	r.Register(clusters.ShadeConfiguration)              // 0x0100
	r.Register(clusters.DoorLock)                        // 0x0101
	r.Register(clusters.WindowCovering)                  // 0x0102
	r.Register(clusters.BarrierControl)                  // 0x0103

	// HVAC (0x0200–0x02FF)
	r.Register(clusters.PumpConfigurationAndControl)     // 0x0200
	r.Register(clusters.Thermostat)                      // 0x0201
	r.Register(clusters.FanControl)                      // 0x0202
	r.Register(clusters.ThermostatUserInterfaceConfiguration) // 0x0204

	// Lighting (0x0300–0x03FF)
	r.Register(clusters.ColorControl)                    // 0x0300
	r.Register(clusters.BallastConfiguration)            // 0x0301

	// Measurement & Sensing (0x0400–0x04FF)
	r.Register(clusters.IlluminanceMeasurement)          // 0x0400
	r.Register(clusters.IlluminanceLevelSensing)         // 0x0401
	r.Register(clusters.TemperatureMeasurement)          // 0x0402
	r.Register(clusters.PressureMeasurement)             // 0x0403
	r.Register(clusters.FlowMeasurement)                 // 0x0404
	r.Register(clusters.RelativeHumidity)                // 0x0405
	r.Register(clusters.OccupancySensing)                // 0x0406
	r.Register(clusters.SoilMoisture)                    // 0x0408
	r.Register(clusters.PHMeasurement)                   // 0x0409
	r.Register(clusters.CarbonMonoxide)                  // 0x040C
	r.Register(clusters.CarbonDioxide)                   // 0x040D
	r.Register(clusters.PM25Measurement)                 // 0x042A
	r.Register(clusters.FormaldehydeMeasurement)         // 0x042B

	// Security & Safety (0x0500–0x05FF)
	r.Register(clusters.IASZone)                         // 0x0500
	r.Register(clusters.IASACE)                          // 0x0501
	r.Register(clusters.IASWD)                           // 0x0502

	// Smart Energy (0x0700–0x07FF)
	r.Register(clusters.Metering)                        // 0x0702

	// Home Automation (0x0B00–0x0BFF)
	r.Register(clusters.ApplianceIdentification)         // 0x0B00
	r.Register(clusters.MeterIdentification)             // 0x0B01
	r.Register(clusters.ApplianceEventsAndAlerts)        // 0x0B02
	r.Register(clusters.ApplianceStatistics)             // 0x0B03
	r.Register(clusters.ElectricalMeasurement)           // 0x0B04
	r.Register(clusters.Diagnostics)                     // 0x0B05

	// Touchlink (0x1000)
	r.Register(clusters.TouchlinkCommissioning)          // 0x1000
}
