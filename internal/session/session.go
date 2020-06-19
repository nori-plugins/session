package session

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	rest "github.com/cheebo/gorest"
	"github.com/dgrijalva/jwt-go"
	"github.com/nori-io/nori-common/endpoint"
	"github.com/nori-io/nori-common/v2/plugin"
	"github.com/nori-io/nori-interfaces/interfaces"

	c "github.com/nori-io/cache-memory/pkg"

)

type Instance struct {
	cache c.Cache
}

func New(r plugin.Registry) (*Instance, error) {
	cache, err := c.GetCache(r)

	if err != nil {
		return nil, err
	}

	return &Instance{
		cache: cache,
	}, nil
}

func (i *Instance) Get(key []byte, data interface{}) error {
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

func (i *Instance) Save(key []byte, data interface{}, exp time.Duration) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	return i.cache.Set(key, buf.Bytes(), exp)
}

func (i *Instance) Delete(key []byte) error {
	return i.cache.Delete(key)
}

func (i *Instance) SessionId(ctx context.Context) []byte {
	str := ctx.Value(interfaces.SessionIdContextKey).(string)
	buf := make([]byte, 0, len(str))
	w := bytes.NewBuffer(buf)
	w.WriteString(str)
	return w.Bytes()
}

func (i *Instance) Verify() endpoint.Middleware {
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

func (i *Instance) verify(key []byte, verify interfaces.SessionVerification) (interface{}, error) {
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
