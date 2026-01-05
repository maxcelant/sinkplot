package store

import (
	"fmt"

	"github.com/maxcelant/sinkplot/internal/config"
	"github.com/maxcelant/sinkplot/pkg/cache"
)

// Store is the in-memory cache for the config objects so they can be quickly found and updated

// COMMENT: I plan on adding some revision tracking here in the future by having a wrapping metadata object
// For now, this works

type Store interface {
	Get(string) (*config.App, error)
	Set(string, config.App)
}

type configStore struct {
	c cache.Cache[*config.App]
}

func New() Store {
	return &configStore{
		c: cache.New[*config.App](),
	}
}

func (cs configStore) Get(name string) (*config.App, error) {
	app, ok := cs.c.Get(name)
	if !ok {
		return nil, fmt.Errorf("failed to find request config object %s", name)
	}
	return app, nil
}

// Set stores a copy of an object. This ensures that mutations to the original after this call
// do _not_ affect the stored object.
func (cs *configStore) Set(name string, obj config.App) {
	cs.c.Set(name, &obj)
}
