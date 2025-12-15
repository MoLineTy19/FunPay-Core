package api

import (
	"FunPay-Core/pkg"
	"FunPay-Core/pkg/types"
	"FunPay-Core/pkg/utils"
	"fmt"
	"strings"
	"time"
)

func DeleteLot(client *pkg.Client, lotId string, title string) ([]byte, error) {
	offer := &types.Offer{
		CSRFToken:     "293ts9v7ab7lg55k",
		FormCreatedAt: fmt.Sprintf("%d", time.Now().Unix()),
		NodeID:        "2043",
		OfferID:       "60201024",
		Price:         "10000",
		Amount:        "100",
		Deleted:       "1",
		Active:        "off",
		Fields: map[string]map[string]string{
			"summary": {
				"ru": "ыыыыы",
				"en": "sssss",
			},
			"desc": {
				"ru": "ыыыыы",
				"en": "sssss",
			},
			"payment_msg": {
				"ru": "ыыыыы",
				"en": "sssss",
			},
		},
	}

	formData := utils.ToURLValues(offer)

	response, err := client.Post("https://funpay.com/lots/offerSave", strings.NewReader(formData.Encode()))

	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return nil, err
	}

	return response, nil
}
