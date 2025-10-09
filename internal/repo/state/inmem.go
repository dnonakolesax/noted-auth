package state

import (
	"errors"
	"time"

	"github.com/muesli/cache2go"

	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

type InMemStateRepo struct {
	client *cache2go.CacheTable
}

func (sr *InMemStateRepo) SetState(state string, redirectURI string, timeout int64) error {
	sr.client.Add(state, time.Second*time.Duration(timeout), redirectURI)

	return nil
}

func (sr *InMemStateRepo) GetState(state string) (string, error) {
	val, err := sr.client.Value(state)

	if err != nil {
		if errors.Is(err, cache2go.ErrKeyNotFound) {
			return "", errorvals.ErrObjectNotFoundInRepoError
		}
		return "", err
	}

	stringData, ok := val.Data().(string)

	if !ok {
		return "", errors.New("proebali")
	}

	return stringData, nil
}
