package rest

import (
	"FunPay-Core/internal/engine"
	"context"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	buf             *engine.Buffer
	token           string
	mux             *http.ServeMux
	state           atomic.Value
	account         atomic.Value
	stateMu         sync.RWMutex
	startedAt       time.Time
	offerCreator    OfferCreator
	offerEditor     OfferEditor
	offerDeleter    OfferDeleter
	offerLister     OfferLister
	offerFormGetter OfferFormGetter
	orderLister     OrderLister
	orderGetter     OrderGetter
	orderRefunder   OrderRefunder
	chatMessager    ChatMessager

	resumeMu sync.RWMutex
	resumeCh chan<- struct{}
}

func NewServer(buf *engine.Buffer, token string) *Server {
	s := &Server{
		buf:       buf,
		token:     token,
		mux:       http.NewServeMux(),
		startedAt: time.Now(),
	}
	s.state.Store("healthy")
	s.account.Store(AccountSnapshot{})
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /events/poll", s.handleEventsPoll)
	s.mux.HandleFunc("GET /account", s.handleAccount)
	s.mux.HandleFunc("POST /offers", s.handleOffersCreate)
	s.mux.HandleFunc("PATCH /offers/{node}/{offer}", s.handleOffersUpdate)
	s.mux.HandleFunc("DELETE /offers/{node}/{offer}", s.handleOffersDelete)
	s.mux.HandleFunc("GET /offers/form", s.handleOffersForm)
	s.mux.HandleFunc("GET /offers/{node}", s.handleOffersList)
	s.mux.HandleFunc("GET /orders", s.handleOrdersList)
	s.mux.HandleFunc("GET /orders/{id}", s.handleOrderDetail)
	s.mux.HandleFunc("POST /orders/{id}/refund", s.handleOrderRefund)
	s.mux.HandleFunc("POST /chats/{id}/messages", s.handleChatMessage)
	s.mux.HandleFunc("POST /control/resume", s.handleControlResume)
	return s
}

func (s *Server) SetState(state string) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.state.Store(state)
}

func (s *Server) SetAccount(a AccountSnapshot) {
	s.account.Store(a)
}

func (s *Server) SetOfferCreator(oc OfferCreator) {
	s.offerCreator = oc
}

func (s *Server) SetOfferEditor(e OfferEditor)          { s.offerEditor = e }
func (s *Server) SetOfferDeleter(d OfferDeleter)        { s.offerDeleter = d }
func (s *Server) SetOfferLister(l OfferLister)          { s.offerLister = l }
func (s *Server) SetOfferFormGetter(fg OfferFormGetter) { s.offerFormGetter = fg }

func (s *Server) SetOrderLister(l OrderLister)     { s.orderLister = l }
func (s *Server) SetOrderGetter(g OrderGetter)     { s.orderGetter = g }
func (s *Server) SetOrderRefunder(r OrderRefunder) { s.orderRefunder = r }
func (s *Server) SetChatMessager(m ChatMessager)   { s.chatMessager = m }

func (s *Server) SetResumeCh(ch chan<- struct{}) {
	s.resumeMu.Lock()
	defer s.resumeMu.Unlock()
	s.resumeCh = ch
}

func (s *Server) Start(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: authMiddleware(s.token, s.mux)}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
