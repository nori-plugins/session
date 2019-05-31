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
	"bytes"
	"context"
	"encoding/gob"
	"time"

	rest "github.com/cheebo/gorest"
	"github.com/dgrijalva/jwt-go"
	"github.com/nori-io/nori-common/logger"

	"github.com/nori-io/nori-common/config"
	"github.com/nori-io/nori-common/endpoint"
	"github.com/nori-io/nori-interfaces/interfaces"
	"github.com/nori-io/nori-common/meta"
	noriPlugin "github.com/nori-io/nori-common/plugin"
)

type plugin struct {
	instance interfaces.Session
	config   *pluginConfig
}

type pluginConfig struct {
	VerificationType config.String
}

type instance struct {
	cache  interfaces.Cache
	config *pluginConfig
	log    logger.Writer
}

var (
	Plugin plugin
)

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
			interfaces.CacheInterface.Dependency("1.0.0"),
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

func (i *instance) Get(key []byte, data interface{}) error {
	val, err := i.cache.Get(key)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	_, err = buf.Write(val)
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(&buf)
	err = dec.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func (i *instance) Save(key []byte, data interface{}, exp time.Duration) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	return i.cache.Set(key, buf.Bytes(), exp)
}

func (i *instance) Delete(key []byte) error {
	return i.cache.Delete(key)
}

func (i *instance) SessionId(ctx context.Context) []byte {
	str := ctx.Value(interfaces.SessionIdContextKey).(string)
	buf := make([]byte, 0, len(str))
	w := bytes.NewBuffer(buf)
	w.WriteString(str)
	return w.Bytes()
}

func (i *instance) Verify() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			var verify interfaces.SessionVerification
			switch i.config.VerificationType() {
			case interfaces.NoVerify.String():
				verify = interfaces.NoVerify
				break
			case interfaces.WhiteList.String():
				verify = interfaces.WhiteList
				break
			case interfaces.BlackList.String():
				verify = interfaces.BlackList
				break
			}

			data := ctx.Value(interfaces.AuthDataContextKey)

			var sid string
			claims, ok := data.(jwt.MapClaims)
			if !ok {
				return nil, rest.ErrorInternal("Internal error")
			} else {
				iid, ok := claims["jti"]
				if !ok {
					return nil, rest.ErrorInternal("Internal error")
				}
				sid, ok = iid.(string)
				if !ok {
					return nil, rest.ErrorInternal("Internal error")
				}
			}

			if verify != interfaces.NoVerify {
				state, err := i.verify([]byte(sid), verify)
				if err != nil {
					return nil, rest.ErrorInternal("Internal error")
				}

				switch state {
				case interfaces.SessionLocked:
					return "", rest.ErrorLocked("Locked")
				case interfaces.SessionError:
					return "", rest.ErrorInternal("Internal error")
				case interfaces.SessionBlocked:
					return "", rest.AccessForbidden()
				case interfaces.SessionExpired:
					return "", rest.AccessForbidden()
				case interfaces.SessionClosed:
					return "", rest.AccessForbidden()
				}
			}

			return next(context.WithValue(ctx, interfaces.SessionIdContextKey, sid), request)
		}
	}
}

func (i *instance) verify(key []byte, verify interfaces.SessionVerification) (interface{}, error) {
	switch verify {
	case interfaces.WhiteList:
		state, err := i.cache.Get(key)
		if err != nil {
			return interfaces.SessionError, err
		}
		return interfaces.State(state), nil
	default:
		return interfaces.SessionActive, nil
	}
}
