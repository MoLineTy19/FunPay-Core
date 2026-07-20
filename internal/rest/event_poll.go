package rest

import (
	"FunPay-Core/internal/engine"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type pollRequest struct {
	Since int64 `json:"since"`
	Wait  int   `json:"wait"`
}

type pollResponse struct {
	Events      []engine.Event `json:"events"`
	NextEventID int64          `json:"nextEventId,omitempty"`
}

func (s *Server) handleEventsPoll(w http.ResponseWriter, r *http.Request) {
	var req pollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, 400, "bad_request", err.Error(), false)
		return
	}

	if req.Since < 0 {
		writeEngineError(w, 400, "bad_request", "since must be >= 0", false)
		return
	}

	if req.Wait < 0 {
		req.Wait = 0
	}
	if req.Wait > 30 {
		req.Wait = 30
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.Wait)*time.Second)
	defer cancel()

	ch := s.buf.Subscribe()
	defer s.buf.Unsubscribe(ch)

	for {
		events, err := s.buf.Since(req.Since)
		if errors.Is(err, engine.ErrCursorTooOld) {
			writeEngineError(w, 409, "cursor_too_old", err.Error(), false)
			return
		}
		if err != nil {
			writeEngineError(w, 500, "internal", err.Error(), true)
			return
		}
		if len(events) > 0 {
			next := events[len(events)-1].EventID
			writeJSON(w, 200, pollResponse{
				Events:      events,
				NextEventID: next,
			})
			return
		}

		select {
		case <-ch:
			continue
		case <-ctx.Done():
			writeJSON(w, 200, pollResponse{
				Events: []engine.Event{},
			})
			return
		}
	}
}
