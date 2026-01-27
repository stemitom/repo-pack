package gh

import (
	"context"
	"errors"
	"math"
	"net"
	"net/http"
	"time"
)

const (
	DefaultMaxRetries = 3
	BaseDelay         = 500 * time.Millisecond
	MaxDelay          = 10 * time.Second
)

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	var opErr *net.OpError
	return errors.As(err, &opErr)
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func backoffDelay(attempt int) time.Duration {
	delay := min(time.Duration(float64(BaseDelay)*math.Pow(2, float64(attempt))), MaxDelay)
	return delay
}

func withRetry[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= DefaultMaxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(attempt - 1)
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		if !isRetryable(lastErr) {
			return result, lastErr
		}
	}

	return result, lastErr
}

func doRequestWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	return withRetry(ctx, func() (*http.Response, error) {
		reqCopy := req.Clone(ctx)
		resp, err := httpClient.Do(reqCopy)
		if err != nil {
			return nil, err
		}

		if isRetryableStatus(resp.StatusCode) {
			resp.Body.Close()
			return nil, &retryableStatusError{StatusCode: resp.StatusCode}
		}

		return resp, nil
	})
}

type retryableStatusError struct {
	StatusCode int
}

func (e *retryableStatusError) Error() string {
	return http.StatusText(e.StatusCode)
}

func (e *retryableStatusError) Timeout() bool {
	return true
}

func (e *retryableStatusError) Temporary() bool {
	return true
}
