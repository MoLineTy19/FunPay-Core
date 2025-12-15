package pkg

import (
	"FunPay-Core/pkg/config"
	"FunPay-Core/pkg/scraper/webpage"
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
		switch resp.StatusCode {
		case 404:
			return nil, fmt.Errorf("страница не найдена! проверьте на правильность ссылки: %s [%s]", url, resp.Status)
		}
		return nil, fmt.Errorf("недопустимый статус: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела запроса: %w", err)
	}

	return body, nil
}

func (c *Client) Post(url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("POST ошибка создания запроса")
	}

	for key, value := range c.headers {
		req.Header.Add(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("недопустимый статус: %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	return responseBody, nil
}

/*
GetLots Собирает активные позиции по лотам
Пример страницы: https://funpay.com/lots/221/
*/
func (c *Client) GetLots(url string) {
	html, err := c.Get(url)
	if err != nil {
		panic(err)
	}
	webpage.GetLotsData(html)
}
