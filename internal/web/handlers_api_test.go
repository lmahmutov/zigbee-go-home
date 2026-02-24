package web

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/ncp"
	"zigbee-go-home/internal/store"
	"zigbee-go-home/internal/zcl"
)

// stubNCP implements ncp.NCP with minimal stubs for testing.
type stubNCP struct {
	permitJoinErr error
	readAttrsResp []ncp.AttributeResponse
	readAttrsErr  error
	sendCmdErr    error
	writeAttrErr  error
}

func (s *stubNCP) Reset(context.Context) error                                { return nil }
func (s *stubNCP) FactoryReset(context.Context) error                         { return nil }
func (s *stubNCP) Init(context.Context) error                                 { return nil }
func (s *stubNCP) FormNetwork(context.Context, ncp.NetworkConfig) error       { return nil }
func (s *stubNCP) StartNetwork(context.Context) error                         { return nil }
func (s *stubNCP) NetworkInfo(context.Context) (*ncp.NetworkInfo, error)       { return nil, nil }
func (s *stubNCP) NetworkScan(context.Context) ([]ncp.NetworkScanResult, error) { return nil, nil }
func (s *stubNCP) GetLocalIEEE(context.Context) ([8]byte, error)              { return [8]byte{}, nil }
func (s *stubNCP) ActiveEndpoints(context.Context, uint16) ([]uint8, error)   { return nil, nil }
func (s *stubNCP) SimpleDescriptor(context.Context, uint16, uint8) (*ncp.SimpleDescriptor, error) {
	return nil, nil
}
func (s *stubNCP) Bind(context.Context, ncp.BindRequest) error              { return nil }
func (s *stubNCP) Unbind(context.Context, ncp.BindRequest) error            { return nil }
func (s *stubNCP) MgmtLeave(context.Context, uint16, [8]byte) error        { return nil }
func (s *stubNCP) OnDeviceJoined(func(ncp.DeviceJoinedEvent))               {}
func (s *stubNCP) OnDeviceLeft(func(ncp.DeviceLeftEvent))                    {}
func (s *stubNCP) OnDeviceAnnounce(func(ncp.DeviceAnnounceEvent))            {}
func (s *stubNCP) OnAttributeReport(func(ncp.AttributeReportEvent))          {}
func (s *stubNCP) GetNCPInfo() *ncp.NCPInfo                                  { return nil }
func (s *stubNCP) Close() error                                              { return nil }

func (s *stubNCP) PermitJoin(_ context.Context, _ uint8) error { return s.permitJoinErr }
func (s *stubNCP) ReadAttributes(_ context.Context, _ ncp.ReadAttributesRequest) ([]ncp.AttributeResponse, error) {
	return s.readAttrsResp, s.readAttrsErr
}
func (s *stubNCP) WriteAttributes(_ context.Context, _ ncp.WriteAttributesRequest) error {
	return s.writeAttrErr
}
func (s *stubNCP) SendCommand(_ context.Context, _ ncp.ClusterCommandRequest) error {
	return s.sendCmdErr
}
func (s *stubNCP) ConfigureReporting(context.Context, ncp.ConfigureReportingRequest) error {
	return nil
}

func setupTestServer(t *testing.T, apiKey string) (*Server, *store.BoltStore, *stubNCP) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	registry := zcl.NewRegistry(logger)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewBoltStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	stub := &stubNCP{}
	events := coordinator.NewEventBus(logger)
	coord := coordinator.New(stub, db, registry, nil, events, coordinator.Config{
		Channel: 15, PanID: 0x1A62,
	}, coordinator.NCPConfig{Type: "nrf52840"}, logger)

	var opts []ServerOption
	if apiKey != "" {
		opts = append(opts, WithAPIKey(apiKey))
	}
	srv, err := NewServer(coord, logger, opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Stop() })

	return srv, db, stub
}

func seedDevice(t *testing.T, db *store.BoltStore, ieee string, short uint16) {
	t.Helper()
	if err := db.SaveDevice(&store.Device{
		IEEEAddress:  ieee,
		ShortAddress: short,
		Manufacturer: "Test",
		Model:        "TestModel",
		Interviewed:  true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAPIListDevices(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)
	seedDevice(t, db, "00158D00012A3B4D", 0x1235)

	req := httptest.NewRequest("GET", "/api/devices", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var devices []store.Device
	if err := json.NewDecoder(w.Body).Decode(&devices); err != nil {
		t.Fatal(err)
	}
	if len(devices) != 2 {
		t.Errorf("device count = %d, want 2", len(devices))
	}
}

func TestAPIGetDevice(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	req := httptest.NewRequest("GET", "/api/devices/00158D00012A3B4C", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var dev store.Device
	if err := json.NewDecoder(w.Body).Decode(&dev); err != nil {
		t.Fatal(err)
	}
	if dev.IEEEAddress != "00158D00012A3B4C" {
		t.Errorf("ieee = %q", dev.IEEEAddress)
	}
}

func TestAPIGetDeviceNotFound(t *testing.T) {
	srv, _, _ := setupTestServer(t, "")

	req := httptest.NewRequest("GET", "/api/devices/FFFFFFFFFFFFFFFF", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAPIDeleteDevice(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	req := httptest.NewRequest("DELETE", "/api/devices/00158D00012A3B4C", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify device is gone.
	_, err := db.GetDevice("00158D00012A3B4C")
	if err == nil {
		t.Error("expected device to be deleted")
	}
}

func TestAPIPermitJoin(t *testing.T) {
	srv, _, _ := setupTestServer(t, "")

	body := `{"duration": 60}`
	req := httptest.NewRequest("POST", "/api/network/permit-join", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["duration"] != "60" {
		t.Errorf("duration = %q, want 60", resp["duration"])
	}
}

func TestAPIReadAttributes(t *testing.T) {
	srv, db, stub := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)
	stub.readAttrsResp = []ncp.AttributeResponse{
		{AttrID: 0, Status: 0, DataType: 0x10, Value: []byte{0x01}},
	}

	body := `{"endpoint": 1, "cluster_id": 6, "attr_ids": [0]}`
	req := httptest.NewRequest("POST", "/api/devices/00158D00012A3B4C/read", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAPIReadAttributesValidation(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	tests := []struct {
		name string
		body string
		want int
	}{
		{"empty attr_ids", `{"endpoint":1,"cluster_id":6,"attr_ids":[]}`, http.StatusBadRequest},
		{"too many attr_ids", `{"endpoint":1,"cluster_id":6,"attr_ids":[` + repeatN("1", 51) + `]}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/devices/00158D00012A3B4C/read", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d, body = %s", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestAPISendCommand(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	body := `{"endpoint": 1, "cluster_id": 6, "command_id": 0}`
	req := httptest.NewRequest("POST", "/api/devices/00158D00012A3B4C/command", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAPISendCommandPayloadLimit(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	// Generate base64 for 129 bytes (base64 of 129 zero bytes).
	// JSON payload field accepts base64-encoded bytes.
	// 129 bytes = "AAAA...AA==" (172 base64 chars).
	payload := make([]byte, 129)
	body, _ := json.Marshal(sendCommandRequest{
		Endpoint:  1,
		ClusterID: 6,
		CommandID: 0,
		Payload:   payload,
	})

	req := httptest.NewRequest("POST", "/api/devices/00158D00012A3B4C/command", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestAPINetworkInfo(t *testing.T) {
	srv, _, _ := setupTestServer(t, "")

	req := httptest.NewRequest("GET", "/api/network", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	if info["channel"] == nil {
		t.Error("expected 'channel' in network info")
	}
}

func TestAPIListClusters(t *testing.T) {
	srv, _, _ := setupTestServer(t, "")

	req := httptest.NewRequest("GET", "/api/clusters", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareHeader(t *testing.T) {
	srv, _, _ := setupTestServer(t, "secret-key")

	// With correct key via header.
	req := httptest.NewRequest("GET", "/api/devices", nil)
	req.Header.Set("X-API-Key", "secret-key")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("correct header key: status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareQueryParam(t *testing.T) {
	srv, _, _ := setupTestServer(t, "secret-key")

	// With correct key via query param.
	req := httptest.NewRequest("GET", "/api/devices?api_key=secret-key", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("correct query key: status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareMissing(t *testing.T) {
	srv, _, _ := setupTestServer(t, "secret-key")

	// Missing key.
	req := httptest.NewRequest("GET", "/api/devices", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("missing key: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareWrongKey(t *testing.T) {
	srv, _, _ := setupTestServer(t, "secret-key")

	req := httptest.NewRequest("GET", "/api/devices", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong key: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAPIRenameDevice(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	body := `{"friendly_name": "Kitchen Light"}`
	req := httptest.NewRequest("PATCH", "/api/devices/00158D00012A3B4C", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["friendly_name"] != "Kitchen Light" {
		t.Errorf("friendly_name = %q, want Kitchen Light", resp["friendly_name"])
	}

	// Verify device was updated in store.
	dev, err := db.GetDevice("00158D00012A3B4C")
	if err != nil {
		t.Fatal(err)
	}
	if dev.FriendlyName != "Kitchen Light" {
		t.Errorf("stored friendly_name = %q, want Kitchen Light", dev.FriendlyName)
	}
}

func TestAPIRenameDeviceNotFound(t *testing.T) {
	srv, _, _ := setupTestServer(t, "")

	body := `{"friendly_name": "Test"}`
	req := httptest.NewRequest("PATCH", "/api/devices/FFFFFFFFFFFFFFFF", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAPIRenameDeviceEmptyName(t *testing.T) {
	srv, db, _ := setupTestServer(t, "")
	seedDevice(t, db, "00158D00012A3B4C", 0x1234)

	body := `{"friendly_name": ""}`
	req := httptest.NewRequest("PATCH", "/api/devices/00158D00012A3B4C", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	dev, err := db.GetDevice("00158D00012A3B4C")
	if err != nil {
		t.Fatal(err)
	}
	if dev.FriendlyName != "" {
		t.Errorf("stored friendly_name = %q, want empty", dev.FriendlyName)
	}
}

// repeatN generates a comma-separated repetition of s, n times.
func repeatN(s string, n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(s)
	}
	return buf.String()
}
