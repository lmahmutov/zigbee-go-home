package web

import (
	"encoding/json"
	"net/http"

	"zigbee-go-home/internal/automation"
)

func (s *Server) handleAutomationsPage(w http.ResponseWriter, r *http.Request) {
	var scripts []*automation.Script
	if s.scriptMgr != nil {
		var err error
		scripts, err = s.scriptMgr.List()
		if err != nil {
			s.logger.Error("list scripts", "err", err)
		}
	}

	s.renderTemplate(w, "automations.html", map[string]interface{}{
		"PageTitle": "Automations",
		"Scripts":   scripts,
	})
}

func (s *Server) handleAPIListAutomations(w http.ResponseWriter, r *http.Request) {
	if s.scriptMgr == nil {
		s.writeJSON(w, http.StatusOK, []interface{}{})
		return
	}
	scripts, err := s.scriptMgr.List()
	if err != nil {
		s.logger.Error("list scripts", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	s.writeJSON(w, http.StatusOK, scripts)
}

func (s *Server) handleAPIGetAutomation(w http.ResponseWriter, r *http.Request) {
	if s.scriptMgr == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	id := r.PathValue("id")
	script, err := s.scriptMgr.Get(id)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "script not found"})
		return
	}
	s.writeJSON(w, http.StatusOK, script)
}

type saveAutomationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LuaCode     string `json:"lua_code"`
	BlocklyXML  string `json:"blockly_xml"`
	Enabled     bool   `json:"enabled"`
}

func (s *Server) handleAPICreateAutomation(w http.ResponseWriter, r *http.Request) {
	if s.scriptMgr == nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "automations not available"})
		return
	}

	var req saveAutomationRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	script := &automation.Script{
		Meta: automation.ScriptMeta{
			Name:        req.Name,
			Description: req.Description,
			Enabled:     req.Enabled,
		},
		LuaCode:    req.LuaCode,
		BlocklyXML: req.BlocklyXML,
	}

	saved, err := s.scriptMgr.Save(script)
	if err != nil {
		s.logger.Error("create script", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if s.autoEngine != nil && saved.Meta.Enabled {
		if err := s.autoEngine.ReloadScript(saved.ID); err != nil {
			s.logger.Error("reload script after create", "id", saved.ID, "err", err)
		}
	}

	s.writeJSON(w, http.StatusCreated, saved)
}

func (s *Server) handleAPIUpdateAutomation(w http.ResponseWriter, r *http.Request) {
	if s.scriptMgr == nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "automations not available"})
		return
	}

	id := r.PathValue("id")
	existing, err := s.scriptMgr.Get(id)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "script not found"})
		return
	}

	var req saveAutomationRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	existing.Meta.Name = req.Name
	existing.Meta.Description = req.Description
	existing.Meta.Enabled = req.Enabled
	existing.LuaCode = req.LuaCode
	existing.BlocklyXML = req.BlocklyXML

	saved, err := s.scriptMgr.Save(existing)
	if err != nil {
		s.logger.Error("update script", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if s.autoEngine != nil {
		if err := s.autoEngine.ReloadScript(saved.ID); err != nil {
			s.logger.Error("reload script after update", "id", saved.ID, "err", err)
		}
	}

	s.writeJSON(w, http.StatusOK, saved)
}

func (s *Server) handleAPIDeleteAutomation(w http.ResponseWriter, r *http.Request) {
	if s.scriptMgr == nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "automations not available"})
		return
	}

	id := r.PathValue("id")
	if s.autoEngine != nil {
		s.autoEngine.StopScript(id)
	}

	if err := s.scriptMgr.Delete(id); err != nil {
		s.logger.Error("delete script", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIRunAutomation(w http.ResponseWriter, r *http.Request) {
	if s.autoEngine == nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "automation engine not available"})
		return
	}

	id := r.PathValue("id")

	// Check if it's a saved script or inline code
	if id == "_inline" {
		// Run inline Lua code from request body
		var req struct {
			LuaCode string `json:"lua_code"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		result := s.autoEngine.RunLuaCode(req.LuaCode)
		s.writeJSON(w, http.StatusOK, result)
		return
	}

	result := s.autoEngine.RunScript(id)
	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAPIToggleAutomation(w http.ResponseWriter, r *http.Request) {
	if s.scriptMgr == nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "automations not available"})
		return
	}

	id := r.PathValue("id")
	script, err := s.scriptMgr.Get(id)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "script not found"})
		return
	}

	script.Meta.Enabled = !script.Meta.Enabled
	saved, err := s.scriptMgr.Save(script)
	if err != nil {
		s.logger.Error("toggle script", "err", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if s.autoEngine != nil {
		if saved.Meta.Enabled {
			if err := s.autoEngine.ReloadScript(saved.ID); err != nil {
				s.logger.Error("reload script after toggle", "id", saved.ID, "err", err)
			}
		} else {
			s.autoEngine.StopScript(saved.ID)
		}
	}

	s.writeJSON(w, http.StatusOK, saved)
}
