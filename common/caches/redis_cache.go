package caches

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	rdb *redis.Client
	ctx context.Context
}

func (p *RedisCache) Init(rdb *redis.Client, ctx context.Context) {
	p.rdb = rdb
	p.ctx = ctx
}

func (p *RedisCache) Get(key string, ptrValue *string) error {
	value, err := p.rdb.Get(p.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}
	*ptrValue = value
	return nil
}

func (p *RedisCache) HSet(key string, field string, value interface{}, expires time.Duration) error {
	if err := p.rdb.HSet(p.ctx, key, field, value).Err(); err != nil {
		return err
	}
	return nil
}

type RedisItemMapGetter map[string]string

func (g RedisItemMapGetter) Get(key string, ptrValue *string) error {
	item, ok := g[key]
	if !ok {
		return ErrCacheMiss
	}
	*ptrValue = item
	return nil
}

func (p *RedisCache) GetMulti(keys ...string) (Getter, error) {
	values, err := p.rdb.MGet(p.ctx, keys...).Result()
	if err != nil {
		return nil, err
	} else if len(values) == 0 {
		return nil, ErrCacheMiss
	}
	m := make(map[string]string)
	for i, key := range keys {
		if i < len(values) && values[i] != nil {
			if s, ok := values[i].(string); ok {
				m[key] = s
			}
		}
	}
	return RedisItemMapGetter(m), nil
}

func (p *RedisCache) Set(key string, value interface{}, expires time.Duration) error {
	if err := p.rdb.Set(p.ctx, key, value, expires).Err(); err != nil {
		return err
	}
	return nil
}

func (p *RedisCache) Delete(key string) error {
	_, err := p.rdb.Del(p.ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func (p *RedisCache) DeleteMulti(keys ...string) error {
	_, err := p.rdb.Del(p.ctx, keys...).Result()
	if err != nil {
		return err
	}
	return nil
}

func (p *RedisCache) exists(key string) (bool, error) {
	exists, err := p.rdb.Exists(p.ctx, key).Result()
	if err != nil {
		return false, err
	}
	if exists == 1 {
		return true, nil
	}
	return false, nil
}

func (p *RedisCache) Add(key string, value interface{}, expires time.Duration) error {
	existed, err := p.exists(key)
	if err != nil {
		return err
	} else if existed {
		return ErrInvalidValue
	}
	return p.Set(key, value, expires)
}

func (p *RedisCache) Replace(key string, value interface{}, expires time.Duration) error {
	existed, err := p.exists(key)
	if err != nil {
		return err
	} else if !existed {
		return ErrNotStored
	}
	return p.Set(key, value, expires)
}
