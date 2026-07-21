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

	if schema.ServerID != "5188" {
		t.Errorf("ServerID: got %q, want 5188", schema.ServerID)
	}

	if schema.CSRFToken == "" {
		t.Errorf("CSRFToken: empty, want non-empty from <input name=csrf_token>")
	}
	if schema.FormCreatedAt == "" {
		t.Errorf("FormCreatedAt: empty, want non-empty from <input name=form_created_at>")
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

func TestParseOfferFormSchemaNoServerID(t *testing.T) {
	body := []byte(`<html><body>
		<div class="lot-fields" data-fields='[{"id":"summary","type":2,"conditions":[]}]'></div>
		</body></html>`)
	schema, err := parseOfferFormSchema(body, "42")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if schema.ServerID != "" {
		t.Errorf("ServerID: got %q, want empty (no select)", schema.ServerID)
	}
	if len(schema.Fields) != 1 {
		t.Errorf("Fields: got %d, want 1", len(schema.Fields))
	}
}

func TestParseOfferFormSchemaCreateForm(t *testing.T) {
	// Живой образец create-формы (/lots/offerEdit?node=791 без offer) — снят с FP.
	// Отличие от edit-формы: нет <option selected> (форма пустая), но csrf_token и
	// form_created_at в hidden <input> присутствуют — session-binding работает.
	body, err := os.ReadFile("../../scratch/offer-edit-create-791.html")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	schema, err := parseOfferFormSchema(body, "791")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if schema.ServerID != "" {
		t.Errorf("ServerID: got %q, want empty (create-форма не имеет selected)", schema.ServerID)
	}
	if schema.CSRFToken == "" {
		t.Errorf("CSRFToken: empty, create-форма обязана содержать hidden csrf_token")
	}
	if schema.FormCreatedAt == "" {
		t.Errorf("FormCreatedAt: empty, create-форма обязана содержать hidden form_created_at")
	}
	if len(schema.Fields) != 6 {
		t.Errorf("Fields: got %d, want 6", len(schema.Fields))
	}
}

func TestParseOfferFormSchemaServersList(t *testing.T) {
	body, err := os.ReadFile("../../scratch/offer-edit-create-791.html")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	schema, err := parseOfferFormSchema(body, "791")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// Создание-форма: нет selected, но все варианты <option> присутствуют.
	if schema.ServerID != "" {
		t.Errorf("ServerID: got %q, want empty (create-форма)", schema.ServerID)
	}
	if len(schema.Servers) != 3 {
		t.Fatalf("Servers count: got %d, want 3 (пустой + Android + iOS)", len(schema.Servers))
	}
	wantServers := map[string]string{
		"":     "\u00a0", // NBSP — пустой option
		"5188": "Android",
		"5187": "iOS",
	}
	got := map[string]string{}
	for _, s := range schema.Servers {
		got[s.Value] = s.Label
	}
	for v, label := range wantServers {
		if got[v] != label {
			t.Errorf("Servers[%q]: got label %q, want %q", v, got[v], label)
		}
	}
}
