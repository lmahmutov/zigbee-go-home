package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) handleAPIListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.coord.Devices().ListDevices()
	if err != nil {
		s.logger.Error("list devices", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	s.writeJSON(w, http.StatusOK, devices)
}

func (s *Server) handleAPIGetDevice(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	dev, err := s.coord.Devices().GetDevice(ieee)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}
	s.writeJSON(w, http.StatusOK, dev)
}

type renameDeviceRequest struct {
	FriendlyName string `json:"friendly_name"`
}

func (s *Server) handleAPIRenameDevice(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	dev, err := s.coord.Devices().GetDevice(ieee)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}

	var req renameDeviceRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	dev.FriendlyName = req.FriendlyName
	if err := s.coord.Devices().SaveDevice(dev); err != nil {
		s.logger.Error("rename device", "err", err, "ieee", ieee)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "friendly_name": dev.FriendlyName})
}

func (s *Server) handleAPIDeleteDevice(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	if err := s.coord.Devices().RemoveDevice(ieee); err != nil {
		s.logger.Error("delete device", "err", err, "ieee", ieee)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type readAttributesRequest struct {
	Endpoint  uint8    `json:"endpoint"`
	ClusterID uint16   `json:"cluster_id"`
	AttrIDs   []uint16 `json:"attr_ids"`
}

func (s *Server) handleAPIReadAttributes(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	dev, err := s.coord.Devices().GetDevice(ieee)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}

	var req readAttributesRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(req.AttrIDs) == 0 {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "attr_ids must not be empty"})
		return
	}
	if len(req.AttrIDs) > 50 {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "attr_ids limited to 50"})
		return
	}

	results, err := s.coord.ReadAttributes(r.Context(), dev.ShortAddress, req.Endpoint, req.ClusterID, req.AttrIDs)
	if err != nil {
		s.logger.Error("read attributes", "err", err, "ieee", ieee)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	s.writeJSON(w, http.StatusOK, results)
}

type writeAttributeRequest struct {
	Endpoint  uint8       `json:"endpoint"`
	ClusterID uint16      `json:"cluster_id"`
	AttrID    uint16      `json:"attr_id"`
	DataType  uint8       `json:"data_type"`
	Value     interface{} `json:"value"`
}

func (s *Server) handleAPIWriteAttribute(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	dev, err := s.coord.Devices().GetDevice(ieee)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}

	var req writeAttributeRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := s.coord.WriteAttribute(r.Context(), dev.ShortAddress, req.Endpoint, req.ClusterID, req.AttrID, req.DataType, req.Value); err != nil {
		s.logger.Error("write attribute", "err", err, "ieee", ieee)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type sendCommandRequest struct {
	Endpoint  uint8  `json:"endpoint"`
	ClusterID uint16 `json:"cluster_id"`
	CommandID uint8  `json:"command_id"`
	Payload   []byte `json:"payload,omitempty"`
}

func (s *Server) handleAPISendCommand(w http.ResponseWriter, r *http.Request) {
	ieee := r.PathValue("ieee")
	dev, err := s.coord.Devices().GetDevice(ieee)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}

	var req sendCommandRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(req.Payload) > 128 {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "payload limited to 128 bytes"})
		return
	}

	if err := s.coord.SendClusterCommand(r.Context(), dev.ShortAddress, req.Endpoint, req.ClusterID, req.CommandID, req.Payload); err != nil {
		s.logger.Error("send command", "err", err, "ieee", ieee)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPINetworkInfo(w http.ResponseWriter, r *http.Request) {
	info := s.coord.NetworkInfo()
	s.writeJSON(w, http.StatusOK, info)
}

type permitJoinRequest struct {
	Duration uint8 `json:"duration"`
}

func (s *Server) handleAPIPermitJoin(w http.ResponseWriter, r *http.Request) {
	var req permitJoinRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := s.coord.PermitJoin(r.Context(), req.Duration); err != nil {
		s.logger.Error("permit join", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"duration": fmt.Sprintf("%d", req.Duration),
	})
}

func (s *Server) handleAPIListClusters(w http.ResponseWriter, r *http.Request) {
	clusters := s.coord.Registry().All()
	s.writeJSON(w, http.StatusOK, clusters)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("writeJSON encode failed", "err", err)
	}
}
