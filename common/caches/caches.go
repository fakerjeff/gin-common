package caches

import (
	"errors"
	"time"
)

type Getter interface {
	// 根据Key获取缓存值，并使ptrValue指向该值
	Get(key string, ptrValue *string) error
}

type CacheProvider interface {
	Getter
	GetMulti(keys ...string) (Getter, error)
	Set(key string, value interface{}, expires time.Duration) error
	Delete(key string) error
	DeleteMulti(keys ...string) error
	Add(key string, value interface{}, expires time.Duration) error
	Replace(key string, value interface{}, expires time.Duration) error
}

var (
	ErrCacheMiss    = errors.New("cache: key not found")
	ErrNotStored    = errors.New("cache: not stored")
	ErrInvalidValue = errors.New("cache: invalid value")
)
