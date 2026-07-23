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
func startTestServerWithBuf(t *testing.T, token string) (*engine.Buffer, string, func()) {
	t.Helper()
	buf := engine.NewBuffer()
	srv := NewServer(buf, token)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = srv.Start(ctx, addr)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)

	return buf, "http://" + addr, func() {
		cancel()
		<-done
	}
}

func startTestServer(t *testing.T, token string) (string, func()) {
	_, url, teardown := startTestServerWithBuf(t, token)
	return url, teardown
}

func TestServerHealthNoToken(t *testing.T) {
	url, teardown := startTestServer(t, "secret")
	defer teardown()

	resp, err := http.Get(url + "/health")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("close body: %v", cerr)
		}
	}()
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("close body: %v", cerr)
		}
	}()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "healthy") {
		t.Fatalf("body: want 'healthy', get %q", body)
	}
}
