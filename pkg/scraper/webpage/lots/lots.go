package lots

import (
	"FunPay-Core/pkg/types"
	"FunPay-Core/pkg/utils"
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetLotsData Извлечение данных из HTML-страницы
func GetLotsData(html []byte) ([]types.Lot, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}

	var lots []types.Lot

	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".tc-server-inside").Text()
		description := s.Find(".tc-desc-text").Text()
		price := utils.ParsePrice(s.Find(".tc-price").Text())

		// Продавец
		sellerUrl, exists := s.Find(".avatar-photo").Attr("data-href")
		if !exists {
			sellerUrl = "-"
		}

		sellerName := s.Find(".media-user-name").Text()
		sellerAge := s.Find(".media-user-info").Text()

		sellerAvatarUrl, _ := s.Find(".avatar-photo").Attr("style")
		sellerAvatarUrl = sellerAvatarUrl[22 : len(sellerAvatarUrl)-1]
		
		sellerReviews := s.Find(".rating-mini-count").Text()
		sellerRating, exists := s.Find(".rating-stars").Attr("class")

		if !exists {
			sellerRating = "Нет отзывов"
		} else {
			if sellerRating != "Нет отзывов" {
				sellerRating = sellerRating[20:21]
			}
		}

		seller := types.Seller{
			Name:      strings.TrimSpace(sellerName),
			Age:       strings.TrimSpace(sellerAge),
			Reviews:   strings.TrimSpace(sellerReviews),
			Rating:    strings.TrimSpace(sellerRating),
			Url:       sellerUrl,
			AvatarUrl: sellerAvatarUrl,
			FromStat:  "lots",
		}

		lot := types.Lot{
			Title:       strings.TrimSpace(title),
			Description: strings.TrimSpace(description),
			Price:       price,
			Seller:      seller,
		}

		lots = append(lots, lot)
	})

	return lots, nil
}
