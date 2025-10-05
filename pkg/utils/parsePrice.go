package utils

import (
	"FunPay-Core/pkg/types"
	"fmt"
	"strconv"
	"strings"
)

// PriceResult содержит цену и валюту

func ParsePrice(priceText string) types.Price {
	if priceText == "" {
		return types.Price{Amount: 0, Currency: "UNKNOWN"}
	}

	// Очищаем строку
	cleaned := strings.TrimSpace(priceText)

	// Извлекаем валюту ДО очистки
	currency := extractCurrency(priceText)

	// Очищаем от текста валют для парсинга числа
	amountText := removeCurrencyText(cleaned)
	amountText = strings.ReplaceAll(amountText, " ", "")
	amountText = normalizeNumberFormat(amountText)

	// Парсим число
	amount, err := strconv.ParseFloat(amountText, 64)
	if err != nil {
		fmt.Printf("Ошибка преобразования цены '%s': %v\n", priceText, err)
		return types.Price{Amount: 0, Currency: currency}
	}

	return types.Price{
		Amount:   amount,
		Currency: currency,
	}
}

// extractCurrency извлекает валюту из текста
func extractCurrency(text string) string {
	text = strings.ToUpper(strings.TrimSpace(text))

	currencyMap := map[string]string{
		"RUB": "RUB", "РУБ": "RUB", "РУБЛЕЙ": "RUB", "₽": "RUB",
		"USD": "USD", "ДОЛЛАР": "USD", "DOLLAR": "USD", "$": "USD",
		"EUR": "EUR", "ЕВРО": "EUR", "EURO": "EUR", "€": "EUR",
	}

	for pattern, currency := range currencyMap {
		if strings.Contains(text, pattern) {
			return currency
		}
	}

	return "UNKNOWN"
}

// removeCurrencyText удаляет текстовые обозначения валют (только для парсинга числа)
func removeCurrencyText(text string) string {
	patterns := []string{
		"RUB", "РУБ", "РУБЛЕЙ", "RUB/РУБЛЕЙ", "₽",
		"USD", "ДОЛЛАР", "DOLLAR", "$",
		"EUR", "ЕВРО", "EURO", "€",
		"/", "\\", "|",
	}

	result := text
	for _, pattern := range patterns {
		result = strings.ReplaceAll(result, pattern, "")
	}

	return strings.TrimSpace(result)
}

// normalizeNumberFormat нормализует числовой формат
func normalizeNumberFormat(text string) string {
	if text == "" {
		return text
	}

	if strings.Contains(text, ",") && !strings.Contains(text, ".") {
		text = strings.ReplaceAll(text, ",", ".")
	} else if strings.Contains(text, ",") && strings.Contains(text, ".") {
		text = strings.ReplaceAll(text, ",", "")
	}

	return text
}
