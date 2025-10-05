package main

import (
	"FunPay-Core/pkg/config"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	client  *http.Client
	headers map[string]string
}

func NewClient() *Client {
	return &Client{
		client:  config.NewHttpClient(),
		headers: make(map[string]string),
	}
}

func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *Client) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("GET ошибка создания запроса: %w", err)
	}

	for key, value := range c.headers {
		req.Header.Add(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("недопустимый статус: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела запроса: %w", err)
	}

	return body, nil
}
