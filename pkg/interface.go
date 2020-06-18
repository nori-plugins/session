package pkg

import (
	"github.com/nori-io/nori-common/v2/errors"
	"github.com/nori-io/nori-common/v2/meta"
	"github.com/nori-io/nori-common/v2/plugin"
)

const (
	SessionInterface meta.Interface = "Session@0.2.0"
)

type Session interface {
	Register(func(*grpc.Server)) // в качестве параметра функции функция и хэндлер реализующий интерфейс grpc сервиса
}

func GetSession(r plugin.Registry) (Session, error) {
	instance, err := r.Interface(SessionInterface)
	if err != nil {
		return nil, err
	}
	i, ok := instance.(Session)
	if !ok {
		return nil, errors.InterfaceAssertError{
			Interface: SessionInterface,
		}
	}
	return i, nil
}
