package session

import (
	"bytes"
	"encoding/gob"
	"time"

	c "github.com/nori-io/interfaces/nori/cache"
)

type Session struct {
	cache c.Cache
}

func (s *Session) Get(key []byte, data interface{}) error {
	value, err := s.cache.Get(key)
	if (value == nil) || (err != nil) {
		return err
	}
	data = value
	return nil
}
func (s *Session) Save(key []byte, data interface{}, exp time.Duration) error {
	bytes, err := getBytes(data)
	if err != nil {
		return err
	}
	return s.cache.Set(key, bytes, exp)
}
func (s *Session) Delete(key []byte) error {
	return s.cache.Delete(key)
}
func (s *Session) IsExists(key []byte) (bool, error) {
	isExist, err := s.cache.Get(key)
	if isExist != nil {
		return false, err
	}
	return true, err
}

func getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
