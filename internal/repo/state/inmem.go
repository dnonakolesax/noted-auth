package state

import (
	"time"

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
		return "", err
	}

	return val.Data().(string), nil
}
