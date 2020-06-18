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

	"github.com/nori-io/nori-common/v2/logger"

	"github.com/nori-io/nori-common/v2/config"
	"github.com/nori-io/nori-common/meta"
	noriPlugin "github.com/nori-io/nori-common/plugin"
	"github.com/nori-io/nori-interfaces/interfaces"

	"github.com/nori-io/session/internal/session"
)

type plugin struct {
	logger   logger.FieldLogger
	instance *session.Instance
	config   *pluginConfig
}

type pluginConfig struct {
	VerificationType config.String
}

func (p *plugin) Init(_ context.Context, config config.Manager) error {
	cm := config.Register(p.Meta())
	p.config = &pluginConfig{
		VerificationType: cm.String("", "verification type: NoVerify, WhiteList or BlackList"),
	}
	return nil
}

func (p *plugin) Instance() interface{} {
	return p.instance
}

func (p plugin) Meta() meta.Meta {
	return &meta.Data{
		ID: meta.ID{
			ID:      "nori/session",
			Version: "1.0.0",
		},
		Author: meta.Author{
			Name: "Nori",
			URI:  "https://nori.io/",
		},
		Core: meta.Core{
			VersionConstraint: ">=1.0.0, <2.0.0",
		},
		Dependencies: []meta.Dependency{
			interfaces.CacheInterface.Dependency(),
		},
		Description: meta.Description{
			Name:        "Nori Session",
			Description: "Nori: Session Interface",
		},
		Interface: interfaces.SessionInterface,
		License: meta.License{
			Title: "",
			Type:  "GPLv3",
			URI:   "https://www.gnu.org/licenses/",
		},
		Tags: []string{"session"},
	}
}

func (p *plugin) Start(ctx context.Context, registry noriPlugin.Registry) error {
	if p.instance == nil {
		cache, _ := interfaces.GetCache(registry)
		instance := &instance{
			cache:  cache,
			config: p.config,
			log:    registry.Logger(p.Meta()),
		}
		p.instance = instance
	}
	return nil
}

func (p *plugin) Stop(_ context.Context, _ noriPlugin.Registry) error {
	p.instance = nil
	return nil
}
