// Copyright (C) 2018 The Nori Authors info@nori.io
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program; if not, see <http://www.gnu.org/licenses/>.
package main

import (
	"context"

	"github.com/nori-io/common/v4/pkg/domain/meta"
	"github.com/nori-io/common/v4/pkg/domain/registry"

	"github.com/nori-io/common/v4/pkg/domain/logger"

	"github.com/nori-io/common/v4/pkg/domain/config"
	p "github.com/nori-io/common/v4/pkg/domain/plugin"
	m "github.com/nori-io/common/v4/pkg/meta"
	c "github.com/nori-io/interfaces/nori/cache"
	s "github.com/nori-io/interfaces/nori/session"

	"github.com/nori-plugins/session/internal/session"
)

func New() p.Plugin {
	return &plugin{}
}

type plugin struct {
	logger   logger.FieldLogger
	instance *session.Instance
	config   conf
}

type conf struct {
	VerificationType config.String
}

func (p plugin) Init(ctx context.Context, config config.Config, log logger.FieldLogger) error {
	p.logger = log
	p.config = conf{
		VerificationType: config.String("", "verification type: NoVerify, WhiteList or BlackList"),
	}
	return nil
}

func (p plugin) Instance() interface{} {
	return p.instance
}

func (p plugin) Meta() meta.Meta {
	return m.Meta{
		ID: m.ID{
			ID:      "nori/session",
			Version: "1.0.0",
		},
		Author: m.Author{
			Name: "Nori",
			URL:  "https://nori.io/",
		},
		Dependencies: []meta.Dependency{
			c.CacheInterface,
		},
		Description: m.Description{
			Title:       "",
			Description: "",
		},
		Interface: s.SessionInterface,
		License: []meta.License{
			m.License{
				Title: "",
				Type:  0,
				URL:   "",
			},
		},
		Links:      nil,
		Repository: nil,
		Tags:       []string{"session"},
	}
}

func (p plugin) Start(ctx context.Context, registry registry.Registry) error {
	var err error
	p.instance, err = session.New(registry)
	if err != nil {
		return err
	}

	return nil
}

func (p plugin) Stop(ctx context.Context, registry registry.Registry) error {
	p.instance = nil
	return nil
}
