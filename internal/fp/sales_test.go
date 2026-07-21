package fp

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseSalesOrders(t *testing.T) {
	body, err := os.ReadFile("../../scratch/orders-trade-list.html")
	if err != nil {
		t.Skipf("sample missing: %v", err)
	}
	orders, err := parseSalesOrders(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("count: got %d, want 1", len(orders))
	}
	o := orders[0]
	if o.ID != "WMBY8JNK" {
		t.Errorf("ID: got %q, want WMBY8JNK", o.ID)
	}
	if o.Status != StatusCancelled {
		t.Errorf("Status: got %q, want cancelled (warning=Возврат)", o.Status)
	}
	if o.BuyerName != "MoLineTy" {
		t.Errorf("BuyerName: got %q, want MoLineTy", o.BuyerName)
	}
	wantPrice, _ := decimal.NewFromString("1.00")
	if !o.Price.Equal(wantPrice) {
		t.Errorf("Price: got %s, want %s", o.Price, wantPrice)
	}
	if o.Summary == "" {
		t.Errorf("Summary empty")
	}
}

func TestParseSalesOrdersEmpty(t *testing.T) {
	body := []byte(`<html><body><div class="content-orders-trade">Нет продаж</div></body></html>`)
	orders, err := parseSalesOrders(body)
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("count: got %d, want 0", len(orders))
	}
}
