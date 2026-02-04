package httpclient

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/metrics"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

// newTestMetrics constructs HTTPRequestMetrics so hc.metrics.* doesn't panic in tests.
// If your metrics struct has different types/constructors, adjust here.
func newTestMetrics() *metrics.HTTPRequestMetrics {
	return &metrics.HTTPRequestMetrics{
		RequestDurations: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "http_request_duration_millisecond_test",
			Help: "test",
		}),
		RequestOks: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_request_ok_test_total",
			Help: "test",
		}),
		RequestBads: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_request_bad_test_total",
			Help: "test",
		}),
		RequestServErrs: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_request_server_err_test_total",
			Help: "test",
		}),
	}
}

func newClientForServer(t *testing.T, srvURL string, rp configs.HTTPRetryPolicyConfig) (*HTTPClient, *atomic.Bool) {
	t.Helper()
	alive := &atomic.Bool{}
	cfg := &configs.HTTPClientConfig{
		DialTimeout:     200 * time.Millisecond,
		KeepAlive:       200 * time.Millisecond,
		MaxIdleConns:    10,
		IdleConnTimeout: 500 * time.Millisecond,
		RequestTimeout:  500 * time.Millisecond,
		RetryPolicy:     rp,
	}
	hc, err := NewWithRetry(srvURL, cfg, newTestMetrics(), alive, testLogger())
	require.NoError(t, err)
	return hc, alive
}

func TestNewWithRetry_SetsAliveTrue_On405Head(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NewWithRetry делает HEAD на endpoint и ждёт 405
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rp := configs.HTTPRetryPolicyConfig{MaxAttempts: 1, RetryOnStatus: map[int]bool{}}
	_, alive := newClientForServer(t, srv.URL, rp)

	require.True(t, alive.Load())
}

func TestNewWithRetry_ReturnsError_OnHeadNot405(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// вместо 405 возвращаем 200 => должно упасть
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	alive := &atomic.Bool{}
	cfg := &configs.HTTPClientConfig{
		DialTimeout:     200 * time.Millisecond,
		KeepAlive:       200 * time.Millisecond,
		MaxIdleConns:    10,
		IdleConnTimeout: 500 * time.Millisecond,
		RequestTimeout:  500 * time.Millisecond,
		RetryPolicy:     configs.HTTPRetryPolicyConfig{MaxAttempts: 1, RetryOnStatus: map[int]bool{}},
	}

	_, err := NewWithRetry(srv.URL, cfg, newTestMetrics(), alive, testLogger())
	require.Error(t, err)
}

func TestCreateRequest_SetsHeadersAndAuth(t *testing.T) {
	t.Parallel()

	hc := &HTTPClient{
		c:        &http.Client{Timeout: 500 * time.Millisecond},
		endpoint: "http://example.com",
		retries:  configs.HTTPRetryPolicyConfig{MaxAttempts: 1, RetryOnStatus: map[int]bool{}},
		metrics:  newTestMetrics(),
		logger:   testLogger(),
		Alive:    &atomic.Bool{},
	}

	req, err := hc.createRequest(context.Background(), http.MethodGet, HTTPRequestParams{
		token:     "tok",
		pathParam: "abc",
	})
	require.NoError(t, err)

	require.Equal(t, "http://example.com/abc", req.URL.String())
	require.Equal(t, "application/json", req.Header.Get(HTTPHeaderContentType))
	require.Equal(t, HTTPAuthorizationPrefix+"tok", req.Header.Get(HTTPHeaderAuthorization))
}

func TestPostForm_SendsURLEncodedBody(t *testing.T) {
	t.Parallel()

	var gotCT string
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		gotCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotBody = string(b)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	hc, _ := newClientForServer(t, srv.URL,
		configs.HTTPRetryPolicyConfig{MaxAttempts: 1, RetryOnStatus: map[int]bool{}})

	form := url.Values{}
	form.Set("a", "1")
	form.Set("b", "hello world")

	resp, err := hc.PostForm(context.Background(), form)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()

	require.Equal(t, HTTPHeaderContentTypeURLEncoded, gotCT)
	require.Contains(t, gotBody, "a=1")
	require.Contains(t, gotBody, "b=hello+world")
}

func TestGet_AddsBearerToken(t *testing.T) {
	t.Parallel()

	var auth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		auth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc, _ := newClientForServer(t, srv.URL,
		configs.HTTPRetryPolicyConfig{MaxAttempts: 1, RetryOnStatus: map[int]bool{}})

	resp, err := hc.Get(context.Background(), "t123")
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()

	require.Equal(t, "Bearer t123", auth)
}

func TestDelete_AppendsPathParam(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotMethod string
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc, _ := newClientForServer(t, srv.URL,
		configs.HTTPRetryPolicyConfig{MaxAttempts: 1, RetryOnStatus: map[int]bool{}})

	resp, err := hc.Delete(context.Background(), "tok", "id-777")
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()

	require.Equal(t, http.MethodDelete, gotMethod)
	require.True(t, strings.HasSuffix(gotPath, "/id-777"))
	require.Equal(t, "Bearer tok", gotAuth)
}

func TestGet_ReturnsError_On4xx(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	hc, _ := newClientForServer(t, srv.URL, configs.HTTPRetryPolicyConfig{
		MaxAttempts:   1,
		RetryOnStatus: map[int]bool{http.StatusUnauthorized: false},
	})

	_, err := hc.Get(context.Background(), "tok")
	require.Error(t, err)
	require.Contains(t, err.Error(), "resp status code: 401")
}
