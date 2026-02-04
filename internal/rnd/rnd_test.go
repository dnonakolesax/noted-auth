package rnd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenRandomString_LengthAndNonNil(t *testing.T) {
	t.Parallel()

	out, err := GenRandomString(0)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Empty(t, out)

	out, err = GenRandomString(32)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, out, 32)
}

func TestGenRandomString_NotAllZeros_ForReasonableLen(t *testing.T) {
	t.Parallel()

	// Это не строгая криптопроверка, просто sanity-check что Read что-то пишет.
	out, err := GenRandomString(64)
	require.NoError(t, err)
	require.Len(t, out, 64)

	require.False(t, bytes.Equal(out, make([]byte, 64)), "unexpected all-zero output")
}

func TestNotSafeGenRandomString_LengthAndAlphabet(t *testing.T) {
	t.Parallel()

	out := NotSafeGenRandomString(0)
	require.NotNil(t, out)
	require.Empty(t, out)

	out = NotSafeGenRandomString(128)
	require.NotNil(t, out)
	require.Len(t, out, 128)

	allowed := make(map[byte]struct{}, len(byteChoice))
	for _, c := range byteChoice {
		allowed[c] = struct{}{}
	}

	for i, b := range out {
		_, ok := allowed[b]
		require.True(t, ok, "byte at index %d (%q) not in allowed alphabet", i, b)
	}
}

func TestNotSafeGenRandomString_NotConstant(t *testing.T) {
	t.Parallel()

	// Вероятность совпадения двух строк длины 32 при равномерном выборе из 62 символов:
	// 62^-32 — практически ноль. В тесте это безопасно.
	a := NotSafeGenRandomString(32)
	b := NotSafeGenRandomString(32)

	require.False(t, bytes.Equal(a, b), "unexpected identical outputs; RNG might be broken")
}
