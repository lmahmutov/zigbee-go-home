package web

import (
	"bytes"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"zigbee-go-home/internal/automation"
	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/store"
	"zigbee-go-home/internal/zcl"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// ServerOption configures the web server.
type ServerOption func(*Server)

// WithAPIKey enables API key authentication.
func WithAPIKey(key string) ServerOption {
	return func(s *Server) {
		s.apiKey = key
	}
}

// WithAllowedOrigins sets allowed WebSocket origin patterns.
func WithAllowedOrigins(origins []string) ServerOption {
	return func(s *Server) {
		s.allowedOrigins = origins
	}
}

// WithDevicesDir sets the path to the devices directory (for serving device images).
func WithDevicesDir(dir string) ServerOption {
	return func(s *Server) {
		s.devicesDir = dir
	}
}

// WithAutomation sets the automation engine and script manager.
func WithAutomation(engine *automation.Engine, mgr *automation.Manager) ServerOption {
	return func(s *Server) {
		s.autoEngine = engine
		s.scriptMgr = mgr
	}
}

// WithVersion sets the application version string shown in the UI.
func WithVersion(v string) ServerOption {
	return func(s *Server) {
		s.version = v
	}
}

// Server is the HTTP server for the web interface.
type Server struct {
	coord          *coordinator.Coordinator
	templates      map[string]*template.Template
	wsHub          *WSHub
	logger         *slog.Logger
	mux            *http.ServeMux
	apiKey         string
	allowedOrigins []string
	devicesDir     string
	photoMu        sync.RWMutex
	photoCache     map[string]string // model -> photo URL (empty string = no photo)
	scriptMgr      *automation.Manager
	autoEngine     *automation.Engine
	version        string
	wg             sync.WaitGroup
	unsubEvents    func()
}

// DeviceView is the enriched view of a device for templates.
type DeviceView struct {
	IEEEAddress     string
	ShortAddress    uint16
	Manufacturer    string
	Model           string
	FriendlyName    string
	Interviewed     bool
	JoinedAt        time.Time
	LastSeen        time.Time
	DeviceType      string // "light", "switch", "sensor", "thermostat", "unknown"
	TypeIcon        string // SVG icon name hint
	HasOnOff        bool
	HasLevel        bool
	OnOffState      string // "on", "off", ""
	EndpointCount   int
	PrimaryEndpoint uint8
	IsKnown         bool   // has a device definition in DeviceDB
	HasPhoto        bool
	PhotoURL        string
	LQI             uint8
	RSSI            int8
	BatteryPercent  int    // -1 if not available
	BatteryVoltage  int    // mV, 0 if not available
	Temperature     int    // 0 if not available
	HasTemperature  bool
	Contact         string // "open", "closed", "" if not available
	Occupancy       string // "detected", "clear", "" if not available
	Illuminance     int    // lux, 0 if not available
	HasIlluminance  bool
	Humidity        int    // 0-100, 0 if not available
	HasHumidity     bool
	LQIQuality      string // "good", "fair", "poor"
	Properties      map[string]any
}

// ClusterInfo is the enriched cluster info for templates.
type ClusterInfo struct {
	ID   uint16
	Name string
}

// EndpointInfo is the enriched endpoint info for templates.
type EndpointInfo struct {
	ID          uint8
	ProfileID   uint16
	DeviceID    uint16
	InClusters  []ClusterInfo
	OutClusters []ClusterInfo
	HasOnOff    bool
	HasLevel    bool
}

// NewServer creates a new web server.
func NewServer(coord *coordinator.Coordinator, logger *slog.Logger, opts ...ServerOption) (*Server, error) {
	// Parse each page template separately with layout to avoid {{define "content"}} conflicts.
	base, err := template.ParseFS(templateFS, "templates/layout.html")
	if err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}
	pages := []string{"index.html", "devices.html", "device_detail.html", "network.html", "automations.html"}
	tmpl := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		cloned, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("clone layout for %s: %w", page, err)
		}
		t, err := cloned.ParseFS(templateFS, "templates/"+page)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", page, err)
		}
		tmpl[page] = t
	}

	s := &Server{
		coord:      coord,
		templates:  tmpl,
		logger:     logger,
		mux:        http.NewServeMux(),
		photoCache: make(map[string]string),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.wsHub = NewWSHub(logger)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.wsHub.Run()
	}()

	// Subscribe to all coordinator events and broadcast via WebSocket
	s.unsubEvents = coord.Events().OnAll(func(event coordinator.Event) {
		s.wsHub.Broadcast(event)
	})

	s.routes()
	return s, nil
}

// Stop gracefully shuts down the WebSocket hub and waits for goroutines.
func (s *Server) Stop() {
	if s.unsubEvents != nil {
		s.unsubEvents()
	}
	s.wsHub.Stop()
	s.wg.Wait()
}

func (s *Server) routes() {
	// Static files
	s.mux.Handle("GET /static/", http.FileServer(http.FS(staticFS)))

	// Device images from devices/img/ directory
	if s.devicesDir != "" {
		imgDir := filepath.Join(s.devicesDir, "img")
		s.mux.Handle("GET /devices/img/", http.StripPrefix("/devices/img/", http.FileServer(http.Dir(imgDir))))
	}

	// HTML pages
	s.mux.HandleFunc("GET /{$}", s.handleIndex)
	s.mux.HandleFunc("GET /devices", s.handleDevicesPage)
	s.mux.HandleFunc("GET /devices/{ieee}", s.handleDeviceDetailPage)
	s.mux.HandleFunc("GET /network", s.handleNetworkPage)

	// REST API
	s.mux.HandleFunc("GET /api/devices", s.handleAPIListDevices)
	s.mux.HandleFunc("GET /api/devices/{ieee}", s.handleAPIGetDevice)
	s.mux.HandleFunc("PATCH /api/devices/{ieee}", s.handleAPIRenameDevice)
	s.mux.HandleFunc("DELETE /api/devices/{ieee}", s.handleAPIDeleteDevice)
	s.mux.HandleFunc("POST /api/devices/{ieee}/read", s.handleAPIReadAttributes)
	s.mux.HandleFunc("POST /api/devices/{ieee}/write", s.handleAPIWriteAttribute)
	s.mux.HandleFunc("POST /api/devices/{ieee}/command", s.handleAPISendCommand)
	s.mux.HandleFunc("GET /api/network", s.handleAPINetworkInfo)
	s.mux.HandleFunc("POST /api/network/permit-join", s.handleAPIPermitJoin)
	s.mux.HandleFunc("GET /api/clusters", s.handleAPIListClusters)
	s.mux.HandleFunc("GET /api/version", s.handleAPIVersion)

	// Automations
	s.mux.HandleFunc("GET /automations", s.handleAutomationsPage)
	s.mux.HandleFunc("GET /api/automations", s.handleAPIListAutomations)
	s.mux.HandleFunc("GET /api/automations/{id}", s.handleAPIGetAutomation)
	s.mux.HandleFunc("POST /api/automations", s.handleAPICreateAutomation)
	s.mux.HandleFunc("PUT /api/automations/{id}", s.handleAPIUpdateAutomation)
	s.mux.HandleFunc("DELETE /api/automations/{id}", s.handleAPIDeleteAutomation)
	s.mux.HandleFunc("POST /api/automations/{id}/toggle", s.handleAPIToggleAutomation)
	s.mux.HandleFunc("POST /api/automations/{id}/run", s.handleAPIRunAutomation)

	// WebSocket
	s.mux.HandleFunc("GET /ws", s.handleWS)
}

// ServeHTTP implements http.Handler, applying auth and CORS middleware.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS: check Origin on mutating requests to prevent CSRF.
	if len(s.allowedOrigins) > 0 {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if r.Method == http.MethodOptions {
				// Preflight request.
				if s.isOriginAllowed(origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
					w.Header().Set("Access-Control-Max-Age", "3600")
					w.WriteHeader(http.StatusNoContent)
					return
				}
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if r.Method != http.MethodGet {
				if !s.isOriginAllowed(origin) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
	}

	if s.apiKey != "" {
		// Only require API key for /api/ endpoints. Static files, HTML pages,
		// WebSocket, and device images are not API-key-protected because
		// browsers cannot send custom headers on page navigation or WS upgrade.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			key := r.Header.Get("X-API-Key")
			if subtle.ConstantTimeCompare([]byte(key), []byte(s.apiKey)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
	}
	s.mux.ServeHTTP(w, r)
}

// isOriginAllowed checks if the origin matches any allowed origin pattern.
func (s *Server) isOriginAllowed(origin string) bool {
	for _, allowed := range s.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// enrichDevice creates a DeviceView from a store.Device.
func (s *Server) enrichDevice(dev *store.Device) DeviceView {
	v := DeviceView{
		IEEEAddress:     dev.IEEEAddress,
		ShortAddress:    dev.ShortAddress,
		Manufacturer:    dev.Manufacturer,
		Model:           dev.Model,
		FriendlyName:    dev.FriendlyName,
		Interviewed:     dev.Interviewed,
		JoinedAt:        dev.JoinedAt,
		LastSeen:        dev.LastSeen,
		DeviceType:      "unknown",
		TypeIcon:        "device",
		EndpointCount:   len(dev.Endpoints),
		PrimaryEndpoint: 1,
		LQI:             dev.LQI,
		RSSI:            dev.RSSI,
		BatteryPercent:  -1,
		Properties:      dev.Properties,
	}

	// LQI quality indicator.
	switch {
	case dev.LQI == 0:
		v.LQIQuality = ""
	case dev.LQI >= 171:
		v.LQIQuality = "good"
	case dev.LQI >= 85:
		v.LQIQuality = "fair"
	default:
		v.LQIQuality = "poor"
	}

	// Check if device has a known definition.
	if db := s.coord.DeviceDB(); db != nil && dev.Manufacturer != "" && dev.Model != "" {
		v.IsKnown = db.Lookup(dev.Manufacturer, dev.Model) != nil
	}

	// Check if device photo exists (cached to avoid repeated stat calls).
	if s.devicesDir != "" && dev.Model != "" {
		s.photoMu.RLock()
		url, ok := s.photoCache[dev.Model]
		s.photoMu.RUnlock()
		if ok {
			if url != "" {
				v.HasPhoto = true
				v.PhotoURL = url
			}
		} else {
			// Sanitize model name for filesystem: spaces → _, * → x, reject path traversal.
			sanitized := strings.ReplaceAll(dev.Model, " ", "_")
			sanitized = strings.ReplaceAll(sanitized, "*", "x")
			if !strings.Contains(sanitized, "..") && !strings.ContainsAny(sanitized, "/\\") {
				for _, ext := range []string{".jpg", ".png", ".webp"} {
					imgPath := filepath.Join(s.devicesDir, "img", sanitized+ext)
					if _, err := os.Stat(imgPath); err == nil {
						url = "/devices/img/" + sanitized + ext
						v.HasPhoto = true
						v.PhotoURL = url
						break
					}
				}
			}
			s.photoMu.Lock()
			// Cap cache at 500 entries to prevent unbounded growth.
			if len(s.photoCache) >= 500 {
				clear(s.photoCache)
			}
			s.photoCache[dev.Model] = url
			s.photoMu.Unlock()
		}
	}

	// Extract known properties.
	if dev.Properties != nil {
		if val, ok := dev.Properties["battery"]; ok {
			if n, ok := toInt(val); ok {
				v.BatteryPercent = n
			}
		}
		if val, ok := dev.Properties["battery_voltage"]; ok {
			if n, ok := toInt(val); ok {
				v.BatteryVoltage = n
			}
		}
		// Check both "temperature" and "device_temperature" (Xiaomi TLV uses "device_temperature").
		for _, key := range []string{"temperature", "device_temperature"} {
			if val, ok := dev.Properties[key]; ok {
				if n, ok := toInt(val); ok {
					v.Temperature = n
					v.HasTemperature = true
					break
				}
			}
		}
		if val, ok := dev.Properties["contact"]; ok {
			if b, ok := val.(bool); ok {
				if b {
					v.Contact = "open"
				} else {
					v.Contact = "closed"
				}
			}
		}
		if val, ok := dev.Properties["occupancy"]; ok {
			if n, ok := toInt(val); ok {
				if n != 0 {
					v.Occupancy = "detected"
				} else {
					v.Occupancy = "clear"
				}
			}
		}
		if val, ok := dev.Properties["illuminance"]; ok {
			if n, ok := toInt(val); ok {
				v.Illuminance = n
				v.HasIlluminance = true
			}
		}
		if val, ok := dev.Properties["humidity"]; ok {
			if n, ok := toInt(val); ok {
				v.Humidity = n
				v.HasHumidity = true
			}
		}
	}

	// Populate on/off state from stored properties.
	if dev.Properties != nil {
		if val, ok := dev.Properties["on_off"]; ok {
			switch b := val.(type) {
			case bool:
				if b {
					v.OnOffState = "on"
				} else {
					v.OnOffState = "off"
				}
			case float64:
				if b != 0 {
					v.OnOffState = "on"
				} else {
					v.OnOffState = "off"
				}
			}
		}
	}

	if len(dev.Endpoints) > 0 {
		v.PrimaryEndpoint = dev.Endpoints[0].ID
	}

	// Determine device type from clusters
	for _, ep := range dev.Endpoints {
		for _, cid := range ep.InClusters {
			switch cid {
			case 0x0006: // On/Off
				v.HasOnOff = true
				if v.DeviceType == "unknown" {
					v.DeviceType = "switch"
					v.TypeIcon = "plug"
				}
			case 0x0008: // Level Control
				v.HasLevel = true
				if v.DeviceType == "switch" || v.DeviceType == "unknown" {
					v.DeviceType = "light"
					v.TypeIcon = "bulb"
				}
			case 0x0300: // Color Control
				v.DeviceType = "light"
				v.TypeIcon = "bulb"
			case 0x0402: // Temperature Measurement
				v.DeviceType = "sensor"
				v.TypeIcon = "thermometer"
			case 0x0405: // Relative Humidity
				if v.DeviceType == "unknown" {
					v.DeviceType = "sensor"
					v.TypeIcon = "humidity"
				}
			case 0x0400: // Illuminance Measurement
				if v.DeviceType == "unknown" {
					v.DeviceType = "sensor"
					v.TypeIcon = "illuminance"
				}
			case 0x0201: // Thermostat
				v.DeviceType = "thermostat"
				v.TypeIcon = "thermostat"
			case 0x0500: // IAS Zone (security sensor)
				if v.DeviceType == "unknown" {
					v.DeviceType = "sensor"
					v.TypeIcon = "security"
				}
			}
		}
	}

	return v
}

// Page handlers
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	devices, err := s.coord.Devices().ListDevices()
	if err != nil {
		s.logger.Error("list devices for index", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var views []DeviceView
	for _, dev := range devices {
		views = append(views, s.enrichDevice(dev))
	}

	s.renderTemplate(w, "index.html", map[string]interface{}{
		"PageTitle":   "Overview",
		"Devices":     views,
		"DeviceCount": len(views),
	})
}

func (s *Server) handleDevicesPage(w http.ResponseWriter, r *http.Request) {
	devices, err := s.coord.Devices().ListDevices()
	if err != nil {
		s.logger.Error("devices page: list devices failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.logger.Debug("devices page", "count", len(devices))

	var views []DeviceView
	for _, dev := range devices {
		views = append(views, s.enrichDevice(dev))
	}

	s.renderTemplate(w, "devices.html", map[string]interface{}{
		"PageTitle": "Devices",
		"Devices":   views,
	})
}

func (s *Server) handleDeviceDetailPage(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	dev, err := s.coord.Devices().GetDevice(ieee)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	view := s.enrichDevice(dev)

	var endpoints []EndpointInfo
	for _, ep := range dev.Endpoints {
		info := EndpointInfo{
			ID:        ep.ID,
			ProfileID: ep.ProfileID,
			DeviceID:  ep.DeviceID,
		}
		for _, cid := range ep.InClusters {
			ci := ClusterInfo{ID: cid}
			if c := s.coord.Registry().Get(cid); c != nil {
				ci.Name = c.Name
			}
			info.InClusters = append(info.InClusters, ci)
			if cid == 0x0006 {
				info.HasOnOff = true
			}
			if cid == 0x0008 {
				info.HasLevel = true
			}
		}
		for _, cid := range ep.OutClusters {
			ci := ClusterInfo{ID: cid}
			if c := s.coord.Registry().Get(cid); c != nil {
				ci.Name = c.Name
			}
			info.OutClusters = append(info.OutClusters, ci)
		}
		endpoints = append(endpoints, info)
	}

	// Build endpoint metadata JSON for action form dropdowns.
	type attrMeta struct {
		ID   uint16 `json:"id"`
		Name string `json:"name"`
		Type uint8  `json:"type"`
	}
	type cmdMeta struct {
		ID   uint8  `json:"id"`
		Name string `json:"name"`
	}
	type clMeta struct {
		ID         uint16     `json:"id"`
		Name       string     `json:"name"`
		Attributes []attrMeta `json:"attributes"`
		Commands   []cmdMeta  `json:"commands"`
	}
	type epMeta struct {
		ID       uint8    `json:"id"`
		Clusters []clMeta `json:"clusters"`
	}

	var epMetas []epMeta
	for _, ep := range dev.Endpoints {
		em := epMeta{ID: ep.ID}
		for _, cid := range ep.InClusters {
			cl := clMeta{ID: cid}
			if def := s.coord.Registry().Get(cid); def != nil {
				cl.Name = def.Name
				for _, a := range def.Attributes {
					cl.Attributes = append(cl.Attributes, attrMeta{
						ID: a.ID, Name: a.Name, Type: a.Type,
					})
				}
				for _, cmd := range def.Commands {
					if cmd.Direction == zcl.DirectionToServer {
						cl.Commands = append(cl.Commands, cmdMeta{
							ID: cmd.ID, Name: cmd.Name,
						})
					}
				}
			}
			em.Clusters = append(em.Clusters, cl)
		}
		epMetas = append(epMetas, em)
	}
	metaJSON, _ := json.Marshal(epMetas)

	pageTitle := dev.IEEEAddress
	if dev.FriendlyName != "" {
		pageTitle = dev.FriendlyName
	} else if dev.Manufacturer != "" && dev.Model != "" {
		pageTitle = dev.Manufacturer + " " + dev.Model
	}

	s.renderTemplate(w, "device_detail.html", map[string]interface{}{
		"PageTitle":    pageTitle,
		"Device":       view,
		"RawDevice":    dev,
		"Endpoints":    endpoints,
		"EndpointMeta": template.JS(metaJSON),
	})
}

func (s *Server) handleNetworkPage(w http.ResponseWriter, r *http.Request) {
	info := s.coord.NetworkInfo()

	devices, _ := s.coord.Devices().ListDevices()
	info["device_count"] = len(devices)
	info["PageTitle"] = "Network"

	s.renderTemplate(w, "network.html", info)
}

// toInt converts various numeric types (including JSON-deserialized float64) to int.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		return int(n), true
	case uint64:
		return int(n), true
	default:
		return 0, false
	}
}

func (s *Server) handleAPIVersion(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"version": s.version})
}

// renderTemplate renders to a buffer first, so partial write failures don't corrupt the response.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	t, ok := s.templates[name]
	if !ok {
		s.logger.Error("template not found", "name", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Inject version and API key into template data if it's a map.
	if m, ok := data.(map[string]interface{}); ok {
		m["Version"] = s.version
		if s.apiKey != "" {
			m["APIKey"] = s.apiKey
		}
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		s.logger.Error("render template", "name", name, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(buf.Bytes()); err != nil {
		s.logger.Debug("write template response", "name", name, "err", err)
	}
}
