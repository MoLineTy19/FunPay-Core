package rest

import "net/http"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	state := s.state.Load().(string)
	writeJSON(w, http.StatusOK, map[string]string{"status": state})
}
