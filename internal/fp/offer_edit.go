package fp

import "errors"

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
