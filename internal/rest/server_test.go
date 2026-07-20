package rest

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"FunPay-Core/internal/engine"
)

// helper: поднимает сервер на свободном порту, возвращает URL + teardown.
func startTestServer(t *testing.T, token string) (string, func()) {
	t.Helper()
	buf := engine.NewBuffer()
	srv := NewServer(buf, token)

	ln, err := net.Listen("tcp", "127.0.0.1:0") // свободный порт
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close() // отдадим addr в http.Server

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = srv.Start(ctx, addr)
		close(done)
	}()
	// дать серверу подняться
	time.Sleep(50 * time.Millisecond)

	return "http://" + addr, func() {
		cancel()
		<-done
	}
}

func TestServerHealthNoToken(t *testing.T) {
	url, teardown := startTestServer(t, "secret")
	defer teardown()

	resp, err := http.Get(url + "/health")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Fatalf("status: got %d, want 401 (no token)", resp.StatusCode)
	}
}

func TestServerHealthWithToken(t *testing.T) {
	url, teardown := startTestServer(t, "secret")
	defer teardown()

	req, _ := http.NewRequest("GET", url+"/health", nil)
	req.Header.Set("X-Engine-Token", "secret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "healthy") {
		t.Fatalf("body: want 'healthy', get %q", body)
	}
}
