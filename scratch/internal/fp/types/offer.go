package types

type Offer struct {
	CSRFToken     string                       `form:"csrf_token"`
	FormCreatedAt string                       `form:"form_created_at"`
	OfferID       string                       `form:"offer_id,omitempty"`  // ID существующего лота (для редактирования)
	NodeID        string                       `form:"node_id"`             // ID раздела
	Deleted       string                       `form:"deleted,omitempty"`   // "1" - удалить лот, "" или "0" - не удалять
	ServerID      string                       `form:"server_id,omitempty"` // ID категории раздела (например, ключи)
	SideID        string                       `form:"side_id,omitempty"`   // ID платформы
	Location      string                       `form:"location,omitempty"`  // ID фракции (Альянс/Орда и т.д.)
	Fields        map[string]map[string]string `form:"fields"`
	Secrets       string                       `form:"secrets,omitempty"`
	Price         string                       `form:"price"`
	Amount        string                       `form:"amount"`
	Active        string                       `form:"active,omitempty"`
}
