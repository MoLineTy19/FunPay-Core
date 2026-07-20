package rest

import (
	"FunPay-Core/internal/engine"
	"context"
	"net/http"
)

type Server struct {
	buf   *engine.Buffer
	token string
	mux   *http.ServeMux
}

func NewServer(buf *engine.Buffer, token string) *Server {
	s := &Server{
		buf:   buf,
		token: token,
		mux:   http.NewServeMux(),
	}
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /events/poll", s.handleEventsPoll)
	return s
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
