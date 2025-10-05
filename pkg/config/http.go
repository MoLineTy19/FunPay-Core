package config

import (
	"net"
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		// Таймаут на запрос
		Timeout: 30 * time.Second,

		// Настройка транспорта
		Transport: &http.Transport{
			
			// Пул соединений
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,

			// Таймауты
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,

			// Настройка TCP соединения
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}
