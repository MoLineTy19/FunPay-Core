package rest

import (
	"FunPay-Core/internal/engine"
	"context"
	"net/http"
	"sync/atomic"
)

type Server struct {
	buf          *engine.Buffer
	token        string
	mux          *http.ServeMux
	state        atomic.Value
	account      atomic.Value
	offerCreator OfferCreator
}

func NewServer(buf *engine.Buffer, token string) *Server {
	s := &Server{
		buf:   buf,
		token: token,
		mux:   http.NewServeMux(),
	}
	s.state.Store("healthy")
	s.account.Store(AccountSnapshot{})
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /events/poll", s.handleEventsPoll)
	s.mux.HandleFunc("GET /account", s.handleAccount)
	s.mux.HandleFunc("POST /offers", s.handleOffersCreate)
	return s
}

func (s *Server) SetState(state string) {
	s.state.Store(state)
}

func (s *Server) SetAccount(a AccountSnapshot) {
	s.account.Store(a)
}

func (s *Server) SetOfferCreator(oc OfferCreator) {
	s.offerCreator = oc
}

func (s *Server) Start(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: authMiddleware(s.token, s.mux)}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
