package tempo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	tempoApiUrl string
	client      *http.Client
}

func NewClient(tempoApiUrl string) *Client {
	return &Client{
		tempoApiUrl: tempoApiUrl,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) QueryTraceByID(ctx context.Context, traceID string) (string, error) {
	baseURL := fmt.Sprintf("%s/api/v2/traces/%s", c.tempoApiUrl, traceID)
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func (c *Client) SearchTraces(ctx context.Context, query string) (string, error) {
	baseURL := fmt.Sprintf("%s/api/search", c.tempoApiUrl)
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("q", query)
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func (c *Client) StatusServices(ctx context.Context) (string, error) {
	baseURL := fmt.Sprintf("%s/status/services", c.tempoApiUrl)
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
