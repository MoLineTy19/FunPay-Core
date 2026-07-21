package fp

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// ErrOfferNotFound — edit-форма лота не найдена (лот не существует / нет доступа).
// Кидается parseOfferEditForm, когда form.form-offer-editor отсутствует в HTML.
var ErrOfferNotFound = errors.New("offer not found")

// LotValues — полный снимок полей edit-формы лота (GET offerEdit?offer=N).
// Берётся как база для encodeOfferEditForm: клиент накладывает изменения поверх.
type LotValues struct {
	NodeID        string
	OfferID       string
	ServerID      string            // из <select name="server_id"> <option selected>
	CSRFToken     string            // из hidden input
	FormCreatedAt string            // из hidden input
	FieldValues   map[string]string // ВСЕ поля формы по имени (fields[level], fields[summary][ru], price, amount, secrets, ...)
	Active        bool              // из checkbox[checked]
	Amount        string            // дубликат для удобства = FieldValues["amount"]
}

// parseOfferEditForm читает ВСЕ input/textarea/select формы form.form-offer-editor
// (как FPC get_lot_fields). Generic — без жёсткой схемы полей.
func parseOfferEditForm(body []byte, nodeID, offerID string) (LotValues, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return LotValues{}, fmt.Errorf("parse html: %w", err)
	}
	form := doc.Find("form.form-offer-editor")
	if form.Length() == 0 {
		return LotValues{}, fmt.Errorf("%w: form.form-offer-editor not found (node=%s offer=%s)", ErrOfferNotFound, nodeID, offerID)
	}

	values := map[string]string{}

	// input (кроме name="query"): value из атрибута, даже если пустой.
	form.Find("input").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || name == "" || name == "query" {
			return
		}
		values[name] = s.AttrOr("value", "")
	})

	// textarea: текст внутри.
	form.Find("textarea").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || name == "" {
			return
		}
		values[name] = s.Text()
	})

	// select: value из <option selected>.
	form.Find("select").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || name == "" {
			return
		}
		sel, _ := s.Find("option[selected]").Attr("value")
		values[name] = sel
	})

	// checkbox[checked]: name → "on" (для Active).
	active := false
	form.Find("input[type=checkbox][checked]").Each(func(_ int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		if name == "active" {
			active = true
		}
		if _, ok := values[name]; !ok {
			values[name] = "on"
		}
	})

	return LotValues{
		NodeID:        nodeID,
		OfferID:       offerID,
		ServerID:      values["server_id"],
		CSRFToken:     values["csrf_token"],
		FormCreatedAt: values["form_created_at"],
		FieldValues:   values,
		Active:        active,
		Amount:        values["amount"],
	}, nil
}
