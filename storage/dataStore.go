package storage

import (
	"strings"

	"github.com/mgloystein/hash_encoder/config"
)

type DataStore interface {
	Get(itemId int) (string, error)
	Reserve() Persistable
}

type Persistable interface {
	ID() int
	Persist(item string) error
}

func NewDataStore(c *config.Config) (DataStore, error) {
	switch strings.ToLower(c.StorageType) {
	// Other things can happen here but I'm just using the memory base storage while libraries
	// outside of stdLib are unavailable.
	// case "sql":
	// 	return nil, nil
	default:
		return newMemoryStorage(), nil
	}
}
