package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"FunPay-Core/internal/fp"

	"github.com/shopspring/decimal"
)

type OfferCreated struct {
	NodeID  string
	OfferID string
	URL     string
}

type OfferCreator interface {
	CreateOffer(ctx context.Context, nodeID, serverID string, fields map[string]string, price decimal.Decimal, amount int, active bool) (OfferCreated, error)
}

type createOfferRequest struct {
	NodeID   string            `json:"nodeId"`
	ServerID string            `json:"serverId"`
	Fields   map[string]string `json:"fields"`
	Price    decimal.Decimal   `json:"price"`
	Amount   int               `json:"amount,omitempty"`
	Active   bool              `json:"active"`
}

type createOfferResponse struct {
	NodeID  string `json:"nodeId"`
	OfferID string `json:"offerId"`
	Created bool   `json:"created"`
	URL     string `json:"url,omitempty"`
}

func (s *Server) handleOffersCreate(w http.ResponseWriter, r *http.Request) {
	var req createOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, http.StatusBadRequest, "bad_request", err.Error(), false)
		return
	}

	if req.NodeID == "" || req.ServerID == "" || len(req.Fields) == 0 || req.Fields["summary"] == "" || req.Price.IsNegative() {
		writeEngineError(w, http.StatusBadRequest, "bad_request",
			"nodeId, serverId, fields.summary required; price must be >= 0", false)
		return
	}

	if s.offerCreator == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable",
			"offer creator not configured", false)
		return
	}

	oc, err := s.offerCreator.CreateOffer(r.Context(), req.NodeID, req.ServerID, req.Fields, req.Price, req.Amount, req.Active)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}

	writeJSON(w, http.StatusCreated, createOfferResponse{
		NodeID:  oc.NodeID,
		OfferID: oc.OfferID,
		Created: true,
		URL:     oc.URL,
	})
}
