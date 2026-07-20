package fp

import (
	"os"
	"testing"
)

func TestParseOfferFormSchema(t *testing.T) {
	body, err := os.ReadFile("../../scratch/offer-edit-form-791.html")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	schema, err := parseOfferFormSchema(body, "791")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if schema.NodeID != "791" {
		t.Errorf("NodeID: got %q, want 791", schema.NodeID)
	}
	if len(schema.Fields) != 6 {
		t.Fatalf("Fields count: got %d, want 6", len(schema.Fields))
	}

	want := map[string]FieldType{
		"level":       FieldText,
		"stage":       FieldText,
		"summary":     FieldMultilingual,
		"desc":        FieldTextarea,
		"payment_msg": FieldTextarea,
		"images":      FieldImages,
	}
	got := map[string]FieldType{}
	for _, f := range schema.Fields {
		got[f.ID] = f.Type
	}
	for id, ft := range want {
		if got[id] != ft {
			t.Errorf("field %q: got type %d, want %d", id, got[id], ft)
		}
	}
}

func TestParseOfferFormSchemaNoFields(t *testing.T) {
	body := []byte(`<html><body><div class="other">no form here</div></body></html>`)
	_, err := parseOfferFormSchema(body, "791")
	if err == nil {
		t.Fatal("want error when .lot-fields not found, got nil")
	}
}
