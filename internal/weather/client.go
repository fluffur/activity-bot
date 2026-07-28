package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	currentURL  = "https://api.weatherapi.com/v1/current.json"
	forecastURL = "https://api.weatherapi.com/v1/forecast.json"
)

type Client struct {
	apiKey string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Current(ctx context.Context, place string) (*CurrentResponse, error) {
	u, err := url.Parse(currentURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("q", place)
	q.Set("lang", "ru")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var apiErr ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)

		if apiErr.Error.Message != "" {
			return nil, fmt.Errorf(apiErr.Error.Message)
		}

		return nil, fmt.Errorf("weather api: %s", resp.Status)
	}

	var out CurrentResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) Forecast(ctx context.Context, place string) (*ForecastResponse, error) {
	u, err := url.Parse(forecastURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("q", place)
	q.Set("days", "3")
	q.Set("lang", "ru")

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr ErrorResponse

		_ = json.NewDecoder(resp.Body).Decode(&apiErr)

		if apiErr.Error.Message != "" {
			return nil, fmt.Errorf(apiErr.Error.Message)
		}

		return nil, fmt.Errorf("weather api: %s", resp.Status)
	}

	var result ForecastResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
