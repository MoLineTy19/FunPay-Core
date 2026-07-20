package rest

import "net/http"

func (s *Server) handleEventsPoll(w http.ResponseWriter, r *http.Request) {
	writeEngineError(w, http.StatusNotImplemented, "not_implemented", "events_poll not yet built", false)
}
