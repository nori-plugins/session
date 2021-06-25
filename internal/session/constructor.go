package session

import (
	"github.com/nori-io/common/v5/pkg/domain/registry"
	c "github.com/nori-io/interfaces/nori/cache"
)

func NewSession(r registry.Registry) (*Session, error) {
	cache, err := c.GetCache(r)

	if err != nil {
		return nil, err
	}

	return &Session{
		cache: cache,
	}, nil
}
