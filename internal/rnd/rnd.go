package rnd

import (
	"crypto/rand"
	"fmt"
	unsafeRand "math/rand/v2"
)

//nolint:gochecknoglobals // нельзя сделать массив константой
var byteChoice = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenRandomString(length uint) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// NotSafeGenRandomString НЕ ИСПОЛЬЗОВАТЬ ТАМ, ГДЕ НУЖЕН КРИПТОСТОЙКИЙ РАНДОМ (НАПРИМЕР, ДЛЯ STATE И PKCE).
func NotSafeGenRandomString(length uint) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = byteChoice[unsafeRand.IntN(len(byteChoice))] //nolint:gosec // не тратимся на сисколы там, где не надо
	}
	return b
}
