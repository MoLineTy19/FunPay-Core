package utils

import (
	"FunPay-Core/pkg/types"
	"fmt"
	"net/url"
)

// ToURLValues преобразует структуру Offer в url.Values для отправки формы на FunPay.
//
// Функция добавляет в параметры только непустые поля структуры, чтобы избежать отправки пустых значений.
// Специальное внимание уделено полю Fields: оно преобразуется в формат fields[name][lang]=value,
// как ожидает сервер FunPay (например: fields[description][ru]=Текст на русском).
func ToURLValues(fp *types.Offer) url.Values {
	data := url.Values{}

	if fp.CSRFToken != "" {
		data.Add("csrf_token", fp.CSRFToken)
	}

	if fp.FormCreatedAt != "" {
		data.Add("form_created_at", fp.FormCreatedAt)
	}

	if fp.OfferID != "" {
		data.Add("offer_id", fp.OfferID)
	}

	if fp.NodeID != "" {
		data.Add("node_id", fp.NodeID)
	}

	if fp.Deleted != "" {
		data.Add("deleted", fp.Deleted)
	}

	if fp.ServerID != "" {
		data.Add("server_id", fp.ServerID)
	}

	if fp.SideID != "" {
		data.Add("side_id", fp.SideID)
	}

	if fp.Location != "" {
		data.Add("location", fp.Location)
	}

	if fp.Secrets != "" {
		data.Add("secrets", fp.Secrets)
	}

	if fp.Price != "" {
		data.Add("price", fp.Price)
	}

	if fp.Amount != "" {
		data.Add("amount", fp.Amount)
	}

	if fp.Active != "" {
		data.Add("active", fp.Active)
	}

	for fieldName, translations := range fp.Fields {
		for lang, value := range translations {
			key := fmt.Sprintf("fields[%s][%s]", fieldName, lang)
			data.Add(key, value)
		}
	}

	return data
}
