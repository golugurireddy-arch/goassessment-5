package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"order-service/middleware"
	"order-service/model"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is OPEN: external product service unavailable")
	ErrNotFound    = errors.New("product not found")
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

type ProductClient interface {
	GetProduct(ctx context.Context, productID string) (*model.Product, error)
}

type ResilientProductClient struct {
	consulAddr  string
	serviceName string
	httpClient  *http.Client
	mu          sync.Mutex
	state       CircuitState
	failures    int
	threshold   int
	cooldown    time.Duration
	lastFail    time.Time
}

func NewResilientProductClient(consulAddr, serviceName string) ProductClient {
	return &ResilientProductClient{
		consulAddr:  consulAddr,
		serviceName: serviceName,
		httpClient:  &http.Client{Timeout: 3 * time.Second},
		threshold:   3,
		cooldown:    10 * time.Second,
		state:       StateClosed,
	}
}

func (c *ResilientProductClient) resolveServiceEndpoint(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/v1/health/service/%s?passing=true", c.consulAddr, c.serviceName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "http://localhost:8081", nil
	}
	defer resp.Body.Close()

	var entries []struct {
		Service struct {
			Address string `json:"Address"`
			Port    int    `json:"Port"`
		} `json:"Service"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil || len(entries) == 0 {
		return "http://localhost:8081", nil
	}

	addr := entries[0].Service.Address
	if addr == "" || addr == "127.0.0.1" {
		addr = "localhost"
	}
	return fmt.Sprintf("http://%s:%d", addr, entries[0].Service.Port), nil
}

func (c *ResilientProductClient) checkCircuit() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateOpen {
		if time.Since(c.lastFail) > c.cooldown {
			c.state = StateHalfOpen
			return true
		}
		return false
	}
	return true
}

func (c *ResilientProductClient) onSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
	c.state = StateClosed
}

func (c *ResilientProductClient) onFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	c.lastFail = time.Now()
	if c.failures >= c.threshold {
		c.state = StateOpen
	}
}

func (c *ResilientProductClient) GetProduct(ctx context.Context, productID string) (*model.Product, error) {
	if !c.checkCircuit() {
		return nil, ErrCircuitOpen
	}

	baseURL, err := c.resolveServiceEndpoint(ctx)
	if err != nil {
		return nil, fmt.Errorf("service discovery failed: %w", err)
	}

	maxRetries := 3
	backoff := 150 * time.Millisecond
	endpoint := fmt.Sprintf("%s/products/%s", baseURL, productID)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}

		if reqID, ok := ctx.Value(middleware.HeaderXRequestID).(string); ok && reqID != "" {
			req.Header.Set(middleware.HeaderXRequestID, reqID)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var p model.Product
				if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
					return nil, fmt.Errorf("failed to decode product response: %w", err)
				}
				c.onSuccess()
				return &p, nil
			}

			if resp.StatusCode == http.StatusNotFound {
				return nil, ErrNotFound
			}

			lastErr = fmt.Errorf("product service returned status %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}

	c.onFailure()
	return nil, fmt.Errorf("external product call failed after %d retries: %w", maxRetries, lastErr)
}
