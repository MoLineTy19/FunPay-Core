package fp

import "testing"

func TestEncodeRefundBody(t *testing.T) {
	body := encodeRefundBody("WMBY8JNK", "tok123")
	want := "csrf_token=tok123&id=WMBY8JNK"
	if body != want {
		t.Errorf("got %q, want %q", body, want)
	}
}
