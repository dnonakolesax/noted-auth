package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const partsInJWT = 3

func ExtractSubject(token string) (string, error) {
	parts := strings.Split(token, ".")

	if len(parts) != partsInJWT {
		return "", errors.New("invalid JWT: not 3 parts")
	}

	middleDecoded, err := base64.RawURLEncoding.DecodeString(parts[1])

	if err != nil {
		return "", fmt.Errorf("error base64 decoding jwt body: %w", err)
	}

	type jwtBody struct {
		Subject string `json:"sub"`
	}

	var body jwtBody
	err = json.Unmarshal(middleDecoded, &body)

	if err != nil {
		return "", fmt.Errorf("error json unmarshaling jwt body: %w", err)
	}

	return body.Subject, nil
}
