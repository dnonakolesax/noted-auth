package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func ExtractSubject(token string) (string, error) {
	parts := strings.Split(token, ".")

	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT: not 3 parts")
	}

	middleDecoded, err := base64.RawURLEncoding.DecodeString(parts[1])

	if err != nil {
		return "", fmt.Errorf("error base64 decoding jwt body: %v", err)
	}

	type jwtBody struct {
		Subject string `json:"sub"`
	}

	var body jwtBody
	err = json.Unmarshal(middleDecoded, &body)

	if err != nil {
		return "", fmt.Errorf("error json unmarshaling jwt body: %v", err)
	}

	return body.Subject, nil
}
