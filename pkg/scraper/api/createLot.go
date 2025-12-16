package api

import (
	"FunPay-Core/pkg"
	"FunPay-Core/pkg/types"
	"FunPay-Core/pkg/utils"
	"fmt"
	"strings"
	"time"
)

// CreateLot создаёт новое предложение (лот) на продажу на FunPay
func CreateLot(
	client *pkg.Client,
	nodeID,
	serverId,
	sideId,
	titleRu,
	titleEn,
	descRu,
	descEn,
	paymentRu,
	paymentEn,
	price,
	amount,
	csrfToken string) ([]byte, error) {

	if csrfToken == "" {
		return nil, fmt.Errorf("csrf_token пустой")
	}
	if nodeID == "" || serverId == "" || sideId == "" {
		return nil, fmt.Errorf("nodeID, serverID или sideID пустые")
	}
	if titleRu == "" || titleEn == "" || descRu == "" || descEn == "" {
		return nil, fmt.Errorf("все текстовые поля обязательны")
	}
	if price == "" || amount == "" {
		return nil, fmt.Errorf("цена и количество обязательны")
	}

	offer := &types.Offer{
		CSRFToken:     csrfToken,
		FormCreatedAt: fmt.Sprintf("%d", time.Now().Unix()),
		NodeID:        nodeID,
		ServerID:      serverId,
		SideID:        sideId,
		Price:         price,
		Amount:        amount,
		Active:        "on",
		Fields: map[string]map[string]string{
			"summary": {
				"ru": titleRu,
				"en": titleEn,
			},
			"desc": {
				"ru": descRu,
				"en": descEn,
			},
			"payment_msg": {
				"ru": paymentRu,
				"en": paymentEn,
			},
		},
	}

	formData := utils.ToURLValues(offer)

	response, err := client.Post("https://funpay.com/lots/offerSave", strings.NewReader(formData.Encode()))

	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса на создание лота: %w", err)

	}

	return response, nil
}
