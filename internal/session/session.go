package session

import (

	"github.com/nori-io/common/v4/pkg/domain/registry"

	c "github.com/nori-io/interfaces/nori/cache"

)

type Instance struct {
	cache c.Cache
}

func New(r registry.Registry) (*Instance, error) {
	cache, err := c.GetCache(r)

	if err != nil {
		return nil, err
	}

	return &Instance{
		cache: cache,
	}, nil
}
