package rest

import (
	"net/http"
)

func (s *Server) handleControlResume(w http.ResponseWriter, r *http.Request) {
	s.stateMu.RLock()
	state := s.state.Load().(string)
	s.stateMu.RUnlock()

	if state != "auth_lost" {
		writeEngineError(w, http.StatusConflict, "conflict",
			"resume is only valid when state=auth_lost (current: "+state+")", false)
		return
	}

	s.resumeMu.Lock()
	defer s.resumeMu.Unlock()
	if s.resumeCh == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable",
			"resume channel not configured", false)
		return
	}

	select {
	case s.resumeCh <- struct{}{}:
	default:
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "accepted",
		"message": "resume signal sent; main will re-read .env and re-init runner",
	})
}
