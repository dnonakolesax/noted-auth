package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

/* ----------------------------- StateRepo stub ----------------------------- */

type stateRepoStub struct {
	mu sync.Mutex

	setCalls []setCall
	get      map[string]getResult
}

type setCall struct {
	state       string
	redirectURI string
	timeout     time.Duration
}

type getResult struct {
	val string
	err error
}

func newStateRepoStub() *stateRepoStub {
	return &stateRepoStub{get: make(map[string]getResult)}
}

func (s *stateRepoStub) SetState(_ context.Context, state string,
	redirectURI string, timeout time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setCalls = append(s.setCalls, setCall{state: state, redirectURI: redirectURI, timeout: timeout})
	return nil
}

func (s *stateRepoStub) GetState(_ context.Context,
	state string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.get[state]
	if !ok {
		return "", errorvals.ErrObjectNotFoundInRepoError
	}
	return r.val, r.err
}

func (s *stateRepoStub) calls() []setCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]setCall, len(s.setCalls))
	copy(out, s.setCalls)
	return out
}

/* ----------------------------- MonitorVault ----------------------------- */

func TestAuthUsecase_MonitorVault_UpdatesClientSecret(t *testing.T) {
	t.Parallel()

	vaultCh := make(chan string, 2)
	defer close(vaultCh)

	ac := &AuthUsecase{
		kcConfig:     configs.KeycloakConfig{ClientSecret: "old"},
		logger:       testLogger(),
		kcCSUpdating: &atomic.Bool{},
	}

	go ac.MonitorVault(vaultCh)

	vaultCh <- "new-secret"
	vaultCh <- "newer-secret"

	require.Eventually(t, func() bool {
		return ac.kcConfig.ClientSecret == "newer-secret"
	}, 500*time.Millisecond, 10*time.Millisecond)
}

/* ----------------------------- GetAuthLink ----------------------------- */

func TestAuthUsecase_GetAuthLink_WritesStateAndReturnsURL(t *testing.T) {
	t.Parallel()

	stateRepo := newStateRepoStub()

	ac := &AuthUsecase{
		authLifetime: 5 * time.Minute,
		repos:        []StateRepo{stateRepo},
		kcConfig: configs.KeycloakConfig{
			RealmAddress:          "https://kc.example",
			AuthEndpoint:          "/auth",
			ClientID:              "cid",
			RedirectURI:           "https://service.example/callback",
			StateLength:           16,
			CodeVerifierLength:    32,
			PostLogoutRedirectURI: "https://service.example/post-logout",
			LogoutEndpoint:        "logout",
		},
		logger: testLogger(),
	}

	ctx := context.Background()
	link, err := ac.GetAuthLink(ctx, "https://return.example/path")
	require.NoError(t, err)

	u, err := url.Parse(link)
	require.NoError(t, err)

	require.Equal(t, "https", u.Scheme)
	require.Equal(t, "kc.example", u.Host)
	require.Equal(t, "/auth", u.Path)

	q := u.Query()
	require.Equal(t, "cid", q.Get("client_id"))
	require.Equal(t, "https://service.example/callback", q.Get("redirect_uri"))
	require.Equal(t, "openid", q.Get("scope"))
	require.Equal(t, "code", q.Get("response_type"))
	require.Equal(t, "S256", q.Get("code_challenge_method"))

	// state должен быть base64url-строкой (RawURLEncoding)
	state := q.Get("state")
	require.NotEmpty(t, state)
	_, err = base64.RawURLEncoding.DecodeString(state)
	require.NoError(t, err)

	// repo должен получить два SetState: state и state:code_verifier
	calls := stateRepo.calls()
	require.Len(t, calls, 2)

	require.Equal(t, state, calls[0].state)
	require.Equal(t, "https://return.example/path", calls[0].redirectURI)
	require.Equal(t, ac.authLifetime, calls[0].timeout)

	require.Equal(t, state+":code_verifier", calls[1].state)
	require.NotEmpty(t, calls[1].redirectURI) // там хранится b64 code_verifier
	require.Equal(t, ac.authLifetime, calls[1].timeout)

	// code_verifier тоже должен быть base64url
	_, err = base64.RawURLEncoding.DecodeString(calls[1].redirectURI)
	require.NoError(t, err)
}

/* ----------------------------- GetToken (pre-HTTP branches) ----------------------------- */

func TestAuthUsecase_GetToken_ReturnURLNotFound(t *testing.T) {
	t.Parallel()

	stateRepo := newStateRepoStub()
	// GetState вернёт ErrObjectNotFoundInRepoError => returnURL останется ""
	// code_verifier тоже пустой
	ac := &AuthUsecase{
		repos:  []StateRepo{stateRepo},
		logger: testLogger(),
		kcConfig: configs.KeycloakConfig{
			StateLength: 16,
		},
		kcCSUpdating: &atomic.Bool{},
	}

	_, err := ac.GetToken(context.Background(), "someState", "code")
	require.Error(t, err)
	require.Equal(t, "return URL not found", err.Error())
}

func TestAuthUsecase_GetToken_CodeVerifierNotFound(t *testing.T) {
	t.Parallel()

	stateRepo := newStateRepoStub()
	stateRepo.get["st"] = getResult{val: "https://return.example", err: nil}
	// st:code_verifier отсутствует => verifier ""
	ac := &AuthUsecase{
		repos:  []StateRepo{stateRepo},
		logger: testLogger(),
		kcConfig: configs.KeycloakConfig{
			StateLength: 16,
		},
		kcCSUpdating: &atomic.Bool{},
	}

	_, err := ac.GetToken(context.Background(), "st", "code")
	require.Error(t, err)
	require.Equal(t, "code verifier not found", err.Error())
}

func TestAuthUsecase_GetToken_RepoErrorPropagates(t *testing.T) {
	t.Parallel()

	stateRepo := newStateRepoStub()
	stateRepo.get["st"] = getResult{val: "", err: io.ErrUnexpectedEOF} // любая ошибка != not found
	ac := &AuthUsecase{
		repos:  []StateRepo{stateRepo},
		logger: testLogger(),
		kcConfig: configs.KeycloakConfig{
			StateLength: 16,
		},
		kcCSUpdating: &atomic.Bool{},
	}

	_, err := ac.GetToken(context.Background(), "st", "code")
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

/* ----------------------------- GetLogoutLink ----------------------------- */

func TestAuthUsecase_GetLogoutLink_OK(t *testing.T) {
	t.Parallel()

	ac := &AuthUsecase{
		kcConfig: configs.KeycloakConfig{
			RealmAddress:          "https://kc.example/realms/r1",
			LogoutEndpoint:        "protocol/openid-connect/logout",
			PostLogoutRedirectURI: "https://service.example/bye",
		},
		logger: testLogger(),
	}

	got := ac.GetLogoutLink(context.Background(), "idtoken123")
	require.Contains(t, got, "post_logout_redirect_uri=")
	require.Contains(t, got, "id_token_hint=idtoken123")
}

/* ----------------------------- GetUserID (http-based, via httptest) ----------------------------- */

func TestAuthUsecase_GetUserID_TokenActive_ReturnsSubjectOnly(t *testing.T) {
	t.Parallel()

	// Keycloak stub
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token/introspect":
			_ = json.NewEncoder(w).Encode(model.IntrospectDTO{Active: true, Subject: "user-123"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ac := &AuthUsecase{
		kcConfig: configs.KeycloakConfig{
			RealmAddress: srv.URL,
			ClientID:     "cid",
			ClientSecret: "sec",
			TokenTimeout: time.Minute,
		},
		kcTimeout:    time.Minute,
		authLifetime: time.Minute,
		logger:       testLogger(),
	}

	out, err := ac.GetUserID(context.Background(), "access", "refresh")
	require.NoError(t, err)
	require.Equal(t, "user-123", out.UserID)
	require.Empty(t, out.AccessToken)
	require.Empty(t, out.RefreshToken)
}

func TestAuthUsecase_GetUserID_TokenInactive_RefreshOK_ReturnsNewTokens(t *testing.T) {
	t.Parallel()

	// Keycloak stub: inactive introspect + refresh returns tokens
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token/introspect":
			_ = json.NewEncoder(w).Encode(model.IntrospectDTO{Active: false, Subject: "user-123"})
		case "/token":
			_ = json.NewEncoder(w).Encode(model.TokenDTO{
				AccessToken:  "newAT",
				RefreshToken: "newRT",
				IDToken:      "newID",
				ExpiresIn:    111,
				RefreshExp:   222,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ac := &AuthUsecase{
		kcConfig: configs.KeycloakConfig{
			RealmAddress: srv.URL,
			ClientID:     "cid",
			ClientSecret: "sec",
			TokenTimeout: time.Minute,
		},
		kcTimeout:    time.Minute,
		authLifetime: time.Minute,
		logger:       testLogger(),
	}

	out, err := ac.GetUserID(context.Background(), "access", "refresh")
	require.NoError(t, err)

	require.Equal(t, "user-123", out.UserID)
	require.Equal(t, "newAT", out.AccessToken)
	require.Equal(t, "newRT", out.RefreshToken)
	require.Equal(t, "newID", out.IDToken)
	require.Equal(t, int64(111), int64(out.ExpiresIn))
	require.Equal(t, int64(222), int64(out.RefreshExp))
}
