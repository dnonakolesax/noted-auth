package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/metrics"
)

type HTTPClient struct {
	c        *http.Client
	endpoint string
	retries  configs.HTTPRetryPolicyConfig
	metrics  *metrics.HTTPRequestMetrics
}

func NewWithRetry(endpoint string, config configs.HTTPClientConfig, reqMetrics *metrics.HTTPRequestMetrics) *HTTPClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConns, // хост один, так что и значение одно
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	return &HTTPClient{
		c:        &http.Client{Transport: tr, Timeout: config.RequestTimeout},
		endpoint: endpoint,
		retries:  config.RetryPolicy,
		metrics:  reqMetrics,
	}
}

func (hc *HTTPClient) observeStatusCode(code int) {
	if code <= http.StatusBadRequest {
		hc.metrics.RequestOks.Inc()
	} else if code <= http.StatusInternalServerError {
		hc.metrics.RequestBads.Inc()
	} else {
		hc.metrics.RequestServErrs.Inc()
	}
}

func (hc *HTTPClient) PostForm(ctx context.Context, form url.Values) (*http.Response, error) {
	encoded := form.Encode()

	var lastErr error
	var resp *http.Response

	for attempt := 1; attempt <= hc.retries.MaxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, hc.endpoint, strings.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ContentLength = int64(len(encoded))

		reqStart := time.Now().UnixMilli()
		resp, err = hc.c.Do(req)
		reqEnd := time.Now().UnixMilli()

		hc.observeStatusCode(resp.StatusCode)

		hc.metrics.RequestDurations.Observe(float64(reqEnd - reqStart))

		if err == nil {
			if !hc.shouldRetryStatus(resp.StatusCode) {
				return resp, nil
			}
			drainAndClose(resp.Body)

			if attempt < hc.retries.MaxAttempts {
				if err = sleepOrCtx(ctx, hc.backoffDelay(attempt)); err != nil {
					return nil, err
				}
			}
			lastErr = fmt.Errorf("retryable HTTP status %d", resp.StatusCode)
			continue
		}

		if !hc.shouldRetryError(err) || attempt == hc.retries.MaxAttempts {
			return nil, err
		}
		lastErr = err
		if err = sleepOrCtx(ctx, hc.backoffDelay(attempt)); err != nil {
			return nil, err
		}
	}

	if resp != nil {
		return resp, lastErr
	}
	return nil, lastErr
}

func (hc *HTTPClient) shouldRetryStatus(code int) bool {
	return hc.retries.RetryOnStatus[code]
}

func (hc *HTTPClient) shouldRetryError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var nerr net.Error
	if errors.As(err, &nerr) {
		return nerr.Timeout()
	}
	return true
}

func (hc *HTTPClient) backoffDelay(attempt int) time.Duration {
	// экспоненциальная задержка: Base * 2^(attempt-1), но не больше MaxDelay
	d := hc.retries.BaseDelay << (attempt - 1)
	if d > hc.retries.MaxDelay {
		d = hc.retries.MaxDelay
	}
	// джиттер ~±20%
	jitterFrac := 0.2
	j := time.Duration(float64(d) * (rand.Float64()*2*jitterFrac - jitterFrac))
	return d + j
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	io.Copy(io.Discard, body) //nolint:errcheck
	body.Close()
}

func sleepOrCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
