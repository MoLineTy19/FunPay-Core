package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"FunPay-Core/internal/engine"
)

// postPoll шлёт POST /events/poll с заданным body и токеном.
func postPoll(t *testing.T, url, token string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url+"/events/poll", bytes.NewReader(b))
	req.Header.Set("X-Engine-Token", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	return resp
}

func TestEventsPollBadCursor(t *testing.T) {
	url, teardown := startTestServer(t, "secret")
	defer teardown()

	resp := postPoll(t, url, "secret", map[string]any{"since": -1, "wait": 0})
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status: got %d, want 400", resp.StatusCode)
	}
}

func TestEventsPollImmediate(t *testing.T) {
	buf, url, teardown := startTestServerWithBuf(t, "secret")
	defer teardown()

	// Пушим событие напрямую в буфер ДО запроса.
	buf.Push([]engine.Event{
		{Type: engine.OrderNew, At: time.Now(), Payload: map[string]any{"id": "ord-1"}},
	})

	resp := postPoll(t, url, "secret", map[string]any{"since": 0, "wait": 0})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var got pollResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Events) != 1 {
		t.Fatalf("events: got %d, want 1", len(got.Events))
	}
	if got.Events[0].Type != engine.OrderNew {
		t.Errorf("type: got %q, want order.new", got.Events[0].Type)
	}
	if got.NextEventID != 1 {
		t.Errorf("nextEventId: got %d, want 1", got.NextEventID)
	}
}

func TestEventsPollCursorTooOld(t *testing.T) {
	buf, url, teardown := startTestServerWithBuf(t, "secret")
	defer teardown()

	// Пушим, потом вытесняем через EvictExpired с далёким now (20 мин > TTL 10 мин).
	buf.Push([]engine.Event{
		{Type: engine.OrderNew, At: time.Now(), Payload: map[string]any{"id": "ord-1"}},
	})
	buf.EvictExpired(time.Now().Add(20 * time.Minute))

	resp := postPoll(t, url, "secret", map[string]any{"since": 0, "wait": 0})
	defer resp.Body.Close()
	if resp.StatusCode != 409 {
		t.Fatalf("status: got %d, want 409", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("cursor_too_old")) {
		t.Fatalf("body: want 'cursor_too_old', got %q", body)
	}
}

func TestEventsPollLongPoll(t *testing.T) {
	buf, url, teardown := startTestServerWithBuf(t, "secret")
	defer teardown()

	start := time.Now()
	// Параллельно: через 200мс пушим событие. Handler к этому моменту
	// должен быть в select, ожидая signal. Если он отдаст раньше 200мс —
	// это не signal path, а баг.
	go func() {
		time.Sleep(200 * time.Millisecond)
		buf.Push([]engine.Event{
			{Type: engine.OrderNew, At: time.Now(), Payload: map[string]any{"id": "ord-1"}},
		})
	}()

	resp := postPoll(t, url, "secret", map[string]any{"since": 0, "wait": 2})
	defer resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode != 200 {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	// Ответ пришёл быстрее 2с — значит signal path сработал, не таймаут.
	if elapsed >= 2*time.Second {
		t.Fatalf("long-poll не разбудился: elapsed %v, want < 2s", elapsed)
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("responded too fast: %v (signal не мог дойти раньше 200мс push'а)", elapsed)
	}
	body, _ := io.ReadAll(resp.Body)
	var got pollResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Events) != 1 {
		t.Fatalf("events: got %d, want 1", len(got.Events))
	}
}

func TestEventsPollTimeout(t *testing.T) {
	url, teardown := startTestServer(t, "secret")
	defer teardown()

	start := time.Now()
	resp := postPoll(t, url, "secret", map[string]any{"since": 0, "wait": 1})
	defer resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode != 200 {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if elapsed < 900*time.Millisecond {
		t.Fatalf("responded too early: %v (want >= 1s)", elapsed)
	}
	body, _ := io.ReadAll(resp.Body)
	var got pollResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Events) != 0 {
		t.Fatalf("events: got %d, want 0", len(got.Events))
	}
}
