package fp

import (
	"os"
	"testing"
)

func TestParseTradeTags(t *testing.T) {
	body, err := os.ReadFile("../../scratch/orders-trade-list.html")
	if err != nil {
		t.Skipf("sample missing: %v", err)
	}
	tags, err := parseTradeTags(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if tags["orders_counters"] != "jhw4vj62" {
		t.Errorf("orders_counters: got %q, want jhw4vj62", tags["orders_counters"])
	}
	if tags["chat_counter"] != "j48eanvw" {
		t.Errorf("chat_counter: got %q, want j48eanvw", tags["chat_counter"])
	}
}

func TestParseTradeTagsMissingElement(t *testing.T) {
	body := []byte(`<html><body>no live-counters here</body></html>`)
	_, err := parseTradeTags(body)
	if err == nil {
		t.Fatal("want error for missing live-counters, got nil")
	}
}

func TestParseTradeTagsMissingAttrs(t *testing.T) {
	body := []byte(`<html><body><div id="live-counters" data-orders="x"></div></body></html>`)
	_, err := parseTradeTags(body)
	if err == nil {
		t.Fatal("want error for missing data-chat attr, got nil")
	}
}
