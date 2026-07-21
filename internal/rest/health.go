package rest

import (
	"net/http"
	"time"
)

type healthResponse struct {
	Status         string `json:"status"`
	Uptime         string `json:"uptime"`
	EventsBuffered int    `json:"eventsBuffered"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	state := s.state.Load().(string)
	resp := healthResponse{
		Status: state,
	}
	if !s.startedAt.IsZero() {
		resp.Uptime = time.Since(s.startedAt).Round(time.Second).String()
	}
	if s.buf != nil {
		resp.EventsBuffered = s.buf.Len()
	}
	writeJSON(w, http.StatusOK, resp)
}
