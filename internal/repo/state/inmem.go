package state

import (
	"errors"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/errorvals"
	"github.com/muesli/cache2go"
)

type InMemStateRepo struct {
	client *cache2go.CacheTable
}

func (sr *InMemStateRepo) SetState(state string, redirectURI string, timeout uint) error {
	sr.client.Add(state, time.Second*time.Duration(timeout), redirectURI)

	return nil
}

func (sr *InMemStateRepo) GetState(state string) (string, error) {
	val, err := sr.client.Value(state)

	if err != nil {
		if errors.As(err, cache2go.ErrKeyNotFound) {
			return "", errorvals.ObjectNotFoundInRepoError
		}
		return "", err
	}

	return val.Data().(string), nil
}
