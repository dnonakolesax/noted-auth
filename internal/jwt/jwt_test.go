package jwt

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func jwtWithPayloadJSON(t *testing.T, payloadJSON string) string {
	t.Helper()

	// header/payload/signature — для ExtractSubject важен только payload (parts[1])
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(payloadJSON))
	sig := "sig" // может быть любым, функция не проверяет подпись

	return strings.Join([]string{header, payload, sig}, ".")
}

func TestExtractSubject_OK(t *testing.T) {
	t.Parallel()

	token := jwtWithPayloadJSON(t, `{"sub":"user-123","exp":1700000000}`)
	sub, err := ExtractSubject(token)
	require.NoError(t, err)
	require.Equal(t, "user-123", sub)
}

func TestExtractSubject_InvalidPartsCount(t *testing.T) {
	t.Parallel()

	_, err := ExtractSubject("a.b") // 2 части
	require.Error(t, err)
	require.Equal(t, "invalid JWT: not 3 parts", err.Error())

	_, err = ExtractSubject("a.b.c.d") // 4 части
	require.Error(t, err)
	require.Equal(t, "invalid JWT: not 3 parts", err.Error())
}

func TestExtractSubject_Base64DecodeError(t *testing.T) {
	t.Parallel()

	// payload не base64url
	token := fmt.Sprintf("hdr.%s.sig", "%%%NOT_BASE64%%%")
	_, err := ExtractSubject(token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error base64 decoding jwt body")
}

func TestExtractSubject_JSONUnmarshalError(t *testing.T) {
	t.Parallel()

	// base64 валидный, но JSON битый
	badJSON := `{"sub":`
	token := jwtWithPayloadJSON(t, badJSON)

	_, err := ExtractSubject(token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error json unmarshaling jwt body")
}

func TestExtractSubject_SubMissing_ReturnsEmptyStringNoError(t *testing.T) {
	t.Parallel()

	token := jwtWithPayloadJSON(t, `{"iss":"x","aud":"y"}`)
	sub, err := ExtractSubject(token)
	require.NoError(t, err)
	require.Empty(t, sub)
}
