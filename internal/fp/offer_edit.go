package fp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
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

	form.Find("input").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || name == "" || name == "query" {
			return
		}
		values[name] = s.AttrOr("value", "")
	})

	form.Find("textarea").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || name == "" {
			return
		}
		values[name] = s.Text()
	})

	form.Find("select").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok || name == "" {
			return
		}
		sel, _ := s.Find("option[selected]").Attr("value")
		values[name] = sel
	})

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

// GetLotFields — GET /lots/offerEdit?node=X&offer=N → parseOfferEditForm.
func (c *Client) GetLotFields(ctx context.Context, nodeID, offerID string) (LotValues, error) {
	url := "https://funpay.com/lots/offerEdit?node=" + nodeID + "&offer=" + offerID
	data, err := c.do(ctx, "GET", url, nil, "")
	if err != nil {
		return LotValues{}, fmt.Errorf("get lot fields: %w", err)
	}
	return parseOfferEditForm(data, nodeID, offerID)
}

// OfferUpdated — результат EditOffer.
type OfferUpdated struct {
	NodeID  string
	OfferID string
	URL     string
}

// OfferDeleted — результат DeleteOffer.
type OfferDeleted struct {
	NodeID  string
	OfferID string
}

// EditOffer — get-then-save. 3 шага:
//  1. GetLotFields(nodeID, offerID) — снимок текущих значений.
//  2. encodeOfferEditForm — накладывает patch (nil → не менять).
//  3. POST /lots/offerSave с referer /lots/offerEdit?node=X&offer=N → parseSaveResponse.
func (c *Client) EditOffer(ctx context.Context, nodeID, offerID string, fields map[string]map[string]string, price *decimal.Decimal, amount *int, active *bool) (OfferUpdated, error) {
	values, err := c.GetLotFields(ctx, nodeID, offerID)
	if err != nil {
		return OfferUpdated{}, err
	}

	form := encodeOfferEditForm(values, fields, price, amount, active)
	referer := "https://funpay.com/lots/offerEdit?node=" + nodeID + "&offer=" + offerID
	body, err := c.doWithReferer(ctx, "POST", "https://funpay.com/lots/offerSave",
		strings.NewReader(form.Encode()),
		"application/x-www-form-urlencoded; charset=UTF-8", referer)
	if err != nil {
		return OfferUpdated{}, fmt.Errorf("offerSave request: %w", err)
	}

	resp, err := parseSaveResponse(body)
	if err != nil {
		return OfferUpdated{}, err // уже ErrAuthLost wrapped
	}
	if !resp.Done || len(resp.Errors) > 0 {
		return OfferUpdated{}, fmt.Errorf("offerSave validation failed: %v", resp.Errors)
	}
	return OfferUpdated{NodeID: nodeID, OfferID: offerID, URL: resp.URL}, nil
}

// DeleteOffer — 2 шага: GetLotFields (для csrf/fca/full-payload) → POST offerSave deleted=1.
func (c *Client) DeleteOffer(ctx context.Context, nodeID, offerID string) (OfferDeleted, error) {
	values, err := c.GetLotFields(ctx, nodeID, offerID)
	if err != nil {
		return OfferDeleted{}, err
	}

	form := encodeDeleteOfferForm(values)
	referer := "https://funpay.com/lots/offerEdit?node=" + nodeID + "&offer=" + offerID
	body, err := c.doWithReferer(ctx, "POST", "https://funpay.com/lots/offerSave",
		strings.NewReader(form.Encode()),
		"application/x-www-form-urlencoded; charset=UTF-8", referer)
	if err != nil {
		return OfferDeleted{}, fmt.Errorf("offerSave request: %w", err)
	}

	resp, err := parseSaveResponse(body)
	if err != nil {
		return OfferDeleted{}, err
	}
	if !resp.Done || len(resp.Errors) > 0 {
		return OfferDeleted{}, fmt.Errorf("offerSave validation failed: %v", resp.Errors)
	}
	return OfferDeleted{NodeID: nodeID, OfferID: offerID}, nil
}
