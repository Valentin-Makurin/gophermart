package accrual

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"strconv"
	"time"

	"github.com/Valentin-Makurin/gophermart/internal/models"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualOrder, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("accrual system not configured")
	}

	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualOrder models.AccrualOrder
		if err := json.NewDecoder(resp.Body).Decode(&accrualOrder); err != nil {
			return nil, err
		}
		return &accrualOrder, nil

	case http.StatusNoContent:
		return nil, fmt.Errorf("order not found in accrual system")

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		seconds, err := strconv.Atoi(retryAfter)
		if err != nil {
			seconds = 60
		}
		return nil, &RateLimitError{RetryAfter: time.Duration(seconds) * time.Second}

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("accrual system internal error")

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after: %v", e.RetryAfter)
}
