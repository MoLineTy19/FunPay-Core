package fp

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseOrderDetail(t *testing.T) {
	body, err := os.ReadFile("../../scratch/orders-trade-with-order.html")
	if err != nil {
		t.Skipf("sample missing: %v", err)
	}
	o, err := parseOrderDetail(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if o.ID != "WMBY8JNK" {
		t.Errorf("ID: got %q, want WMBY8JNK", o.ID)
	}
	if o.BuyerID != 4759067 {
		t.Errorf("BuyerID: got %d, want 4759067", o.BuyerID)
	}
	if o.BuyerName != "MoLineTy" {
		t.Errorf("BuyerName: got %q, want MoLineTy", o.BuyerName)
	}
	if o.ChatID != "users-4759067-16950672" {
		t.Errorf("ChatID: got %q, want users-4759067-16950672", o.ChatID)
	}
	if o.NodeID != "791" {
		t.Errorf("NodeID: got %q, want 791", o.NodeID)
	}
	wantAmount, _ := decimal.NewFromString("1")
	if !o.Amount.Equal(wantAmount) {
		t.Errorf("Amount: got %s, want 1", o.Amount)
	}
	if o.Currency == "" {
		t.Errorf("Currency empty")
	}
}

func TestParseOrderDetailNotFound(t *testing.T) {
	body := []byte(`<html><head><title>Страница не найдена</title></head><body>404</body></html>`)
	_, err := parseOrderDetail(body)
	if err != ErrOrderNotFound {
		t.Fatalf("err: got %v, want ErrOrderNotFound", err)
	}
}
