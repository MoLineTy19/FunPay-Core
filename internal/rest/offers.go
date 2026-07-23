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
	CreateOffer(ctx context.Context, nodeID, serverID string, fields map[string]map[string]string, price decimal.Decimal, amount int, active bool) (OfferCreated, error)
}

// createOfferRequest — тело POST /offers.
type createOfferRequest struct {
	NodeID   string                       `json:"nodeId" example:"80"`
	ServerID string                       `json:"serverId" example:"7"`
	Fields   map[string]map[string]string `json:"fields"`
	Price    decimal.Decimal              `json:"price" example:"150.00"`
	Amount   int                          `json:"amount,omitempty" example:"0"`
	Active   bool                         `json:"active" example:"true"`
}

type createOfferResponse struct {
	NodeID  string `json:"nodeId"`
	OfferID string `json:"offerId"`
	Created bool   `json:"created"`
	URL     string `json:"url,omitempty"`
}

// @Summary      Создать лот
// @Description  Создаёт новый лот в разделе nodeId. Обязательны nodeId, serverId, непустой fields.summary и price >= 0.
// @Tags         offers
// @Accept       json
// @Produce      json
// @Param        request  body      createOfferRequest  true  "Параметры создаваемого лота"
// @Success      201      {object}  createOfferResponse
// @Failure      400      {object}  EngineError  "bad_request"
// @Failure      401      {object}  EngineError  "missing or invalid token"
// @Failure      503      {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500      {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /offers [post]
func (s *Server) handleOffersCreate(w http.ResponseWriter, r *http.Request) {
	var req createOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, http.StatusBadRequest, "bad_request", err.Error(), false)
		return
	}

	if req.NodeID == "" || req.ServerID == "" || len(req.Fields) == 0 || !hasNonEmptyLocale(req.Fields["summary"]) || req.Price.IsNegative() {
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

// --- Offer edit / delete / list / form ---

type OfferEditor interface {
	EditOffer(ctx context.Context, nodeID, offerID string, fields map[string]map[string]string, price *decimal.Decimal, amount *int, active *bool) (OfferUpdated, error)
}

type OfferDeleter interface {
	DeleteOffer(ctx context.Context, nodeID, offerID string) (OfferDeleted, error)
}

type OfferLister interface {
	ListOffers(ctx context.Context, nodeID string) ([]OfferListItem, error)
}

type OfferFormGetter interface {
	GetOfferForm(ctx context.Context, nodeID string) (OfferForm, error)
}

type OfferUpdated struct {
	NodeID  string
	OfferID string
	URL     string
}

type OfferDeleted struct {
	NodeID  string
	OfferID string
}

type OfferListItem struct {
	OfferID string          `json:"offerId"`
	Summary string          `json:"summary"`
	Server  string          `json:"server,omitempty"`
	Amount  string          `json:"amount,omitempty"`
	Price   decimal.Decimal `json:"price"`
}

type OfferForm struct {
	NodeID   string           `json:"nodeId"`
	ServerID string           `json:"serverId,omitempty"`
	Fields   []OfferFormField `json:"fields"`
	Servers  []OfferServer    `json:"servers,omitempty"`
}

type OfferFormField struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}

type OfferServer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type patchOfferRequest struct {
	Fields map[string]map[string]string `json:"fields,omitempty"`
	Price  *decimal.Decimal             `json:"price,omitempty"`
	Amount *int                         `json:"amount,omitempty"`
	Active *bool                        `json:"active,omitempty"`
}

type updateOfferResponse struct {
	NodeID  string `json:"nodeId"`
	OfferID string `json:"offerId"`
	Updated bool   `json:"updated"`
	URL     string `json:"url,omitempty"`
}

// @Summary      Редактировать лот
// @Description  Частичное редактирование лота. В теле передаются только изменяемые поля (fields/price/amount/active). Запрос без единого поля даёт 400.
// @Tags         offers
// @Accept       json
// @Produce      json
// @Param        node     path      string             true  "ID раздела FunPay"
// @Param        offer    path      string             true  "ID лота внутри раздела"
// @Param        request  body      patchOfferRequest  true  "Изменяемые поля лота"
// @Success      200      {object}  updateOfferResponse
// @Failure      400      {object}  EngineError  "bad_request"
// @Failure      401      {object}  EngineError  "missing or invalid token"
// @Failure      404      {object}  EngineError  "offer_not_found"
// @Failure      503      {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500      {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /offers/{node}/{offer} [patch]
func (s *Server) handleOffersUpdate(w http.ResponseWriter, r *http.Request) {
	node := r.PathValue("node")
	offer := r.PathValue("offer")
	if node == "" || offer == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "node and offer path segments required", false)
		return
	}

	var req patchOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, http.StatusBadRequest, "bad_request", err.Error(), false)
		return
	}
	if len(req.Fields) == 0 && req.Price == nil && req.Amount == nil && req.Active == nil {
		writeEngineError(w, http.StatusBadRequest, "bad_request",
			"nothing to update: provide at least one of fields/price/amount/active", false)
		return
	}
	if req.Price != nil && req.Price.IsNegative() {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "price must be >= 0", false)
		return
	}
	if s.offerEditor == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "offer editor not configured", false)
		return
	}

	ou, err := s.offerEditor.EditOffer(r.Context(), node, offer, req.Fields, req.Price, req.Amount, req.Active)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrOfferNotFound) {
			writeEngineError(w, http.StatusNotFound, "offer_not_found", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}

	writeJSON(w, http.StatusOK, updateOfferResponse{NodeID: ou.NodeID, OfferID: ou.OfferID, Updated: true, URL: ou.URL})
}

type deleteOfferResponse struct {
	NodeID  string `json:"nodeId"`
	OfferID string `json:"offerId"`
	Deleted bool   `json:"deleted"`
}

// @Summary      Удалить лот
// @Description  Безвозвратно удаляет лот из раздела.
// @Tags         offers
// @Produce      json
// @Param        node   path      string  true  "ID раздела FunPay"
// @Param        offer  path      string  true  "ID лота внутри раздела"
// @Success      200    {object}  deleteOfferResponse
// @Failure      400    {object}  EngineError  "bad_request"
// @Failure      401    {object}  EngineError  "missing or invalid token"
// @Failure      404    {object}  EngineError  "offer_not_found"
// @Failure      503    {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500    {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /offers/{node}/{offer} [delete]
func (s *Server) handleOffersDelete(w http.ResponseWriter, r *http.Request) {
	node := r.PathValue("node")
	offer := r.PathValue("offer")
	if node == "" || offer == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "node and offer path segments required", false)
		return
	}
	if s.offerDeleter == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "offer deleter not configured", false)
		return
	}

	od, err := s.offerDeleter.DeleteOffer(r.Context(), node, offer)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrOfferNotFound) {
			writeEngineError(w, http.StatusNotFound, "offer_not_found", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}

	writeJSON(w, http.StatusOK, deleteOfferResponse{NodeID: od.NodeID, OfferID: od.OfferID, Deleted: true})
}

type listOffersResponse struct {
	NodeID string          `json:"nodeId"`
	Offers []OfferListItem `json:"offers"`
}

// @Summary      Список лотов в разделе
// @Description  Возвращает все лоты продавца в разделе node. Пустой список — [] (не null).
// @Tags         offers
// @Produce      json
// @Param        node  path      string  true  "ID раздела FunPay"
// @Success      200   {object}  listOffersResponse
// @Failure      400   {object}  EngineError  "bad_request"
// @Failure      401   {object}  EngineError  "missing or invalid token"
// @Failure      503   {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500   {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /offers/{node} [get]
func (s *Server) handleOffersList(w http.ResponseWriter, r *http.Request) {
	node := r.PathValue("node")
	if node == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "node path segment required", false)
		return
	}
	if s.offerLister == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "offer lister not configured", false)
		return
	}

	items, err := s.offerLister.ListOffers(r.Context(), node)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}
	if items == nil {
		items = []OfferListItem{} // не null в JSON
	}

	writeJSON(w, http.StatusOK, listOffersResponse{NodeID: node, Offers: items})
}

// @Summary      Схема формы создания лота
// @Description  Возвращает поля формы (с типами) и список доступных серверов для раздела node. Нужна, чтобы заранее знать, какие fields и serverId подставить в POST /offers.
// @Tags         offers
// @Produce      json
// @Param        node  query     string  true  "ID раздела FunPay"
// @Success      200   {object}  OfferForm
// @Failure      400   {object}  EngineError  "bad_request"
// @Failure      401   {object}  EngineError  "missing or invalid token"
// @Failure      503   {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500   {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /offers/form [get]
func (s *Server) handleOffersForm(w http.ResponseWriter, r *http.Request) {
	node := r.URL.Query().Get("node")
	if node == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "node query param required", false)
		return
	}
	if s.offerFormGetter == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "offer form getter not configured", false)
		return
	}

	form, err := s.offerFormGetter.GetOfferForm(r.Context(), node)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}

	writeJSON(w, http.StatusOK, form)
}

func hasNonEmptyLocale(vals map[string]string) bool {
	for _, v := range vals {
		if v != "" {
			return true
		}
	}
	return false
}
