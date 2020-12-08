package caches

import (
	"errors"
	"time"

	"com.github.gin-common/util"

	"com.github.gin-common/common/bloomfilter"
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

type cacheOption struct {
	key               string               // 缓存key
	bloomFilterOption bloomfilter.BFOption // 布隆过滤器配置
	cacheProvider     CacheProvider        // 缓存Provider
	expires           time.Duration        // 缓存过期时间
	serializer        util.Serializer      //缓存值序列化器
	sync              bool                 //缓存值更新否同步
}

func (option *cacheOption) WithOption(opts ...CacheOptions) {
	for _, opt := range opts {
		opt.apply(option)
	}
}

func newCacheOption(opts ...CacheOptions) *cacheOption {
	option := new(cacheOption)
	option.WithOption(opts...)
	return option
}

type CacheKeyOption string

func (k CacheKeyOption) apply(cacheOption *cacheOption) {
	cacheOption.key = string(k)
}

type CacheBloomFilterOption bloomfilter.BFOption

func (b CacheBloomFilterOption) apply(cacheOption *cacheOption) {
	cacheOption.bloomFilterOption = bloomfilter.BFOption(b)
}

type RedisCacheProvideOption RedisCache

func (c RedisCacheProvideOption) apply(cacheOption *cacheOption) {
	provider := RedisCache(c)
	cacheOption.cacheProvider = &provider
}

type CacheExpiresOption time.Duration

func (e CacheExpiresOption) apply(cacheOption *cacheOption) {
	cacheOption.expires = time.Duration(e)
}

type SerializerOption util.SerializeTool

func (s SerializerOption) apply(cacheOption *cacheOption) {
	cacheOption.serializer = util.SerializeTool(s)
}

type CacheSyncOption bool

func (s CacheSyncOption) apply(cacheOption *cacheOption) {
	cacheOption.sync = bool(s)
}

type CacheOptions interface {
	apply(cacheOption *cacheOption)
}

func CacheEnable(process func() (interface{}, error), condition func() bool, ptr interface{}, opts ...CacheOptions) (r interface{}, e error) {
	// 缓存装饰方法
	// 缓存准入条件
	if !condition() {
		r, e = process()
		return
	}
	options := newCacheOption(opts...)
	key := options.key
	// 若没配置缓存key则直接返回原函数结果
	if key == "" {
		r, e = process()
		return
	}
	// 判断是否开启布隆过滤器(防穿透)
	// 若开启则先检验key是否存在于布隆过滤器中，若不存在直接返回异常
	if options.bloomFilterOption.Enable() {
		// 若开启了布隆过滤器,但并未指定布隆过滤器，则直接返回原函数结果
		filter := options.bloomFilterOption.Filter()
		if filter == nil {
			return process()
		}
		exists, err := filter.Exists(options.bloomFilterOption.Key(), key)
		if err != nil {
			e = err
			return
		}
		if !exists {
			e = errors.New("item is not found")
			return
		}
	}

	// Todo： 此过程(更新缓存)考虑加锁，因为在多goroutine或分布式环境下同一时间相同key可能出现多次计算（分布式锁）（实现方式建议：zk或etcd）
	var result string
	e = options.cacheProvider.Get(key, &result)
	if e != nil {
		return
	}

	if result == "" {
		r, e = process()
		if e != nil {
			return
		}
		e = options.cacheProvider.Set(key, r, options.expires)
		if e != nil {
			return
		}
	}
	// 此处需要将查出来的结果反序列化
	e = options.serializer.Deserialize([]byte(result), ptr)
	if e != nil {
		return
	}
	r = ptr
	return
}

func CachePut(process func() (interface{}, error), condition func() bool, opts ...CacheOptions) (r interface{}, e error) {
	r, e = process()

	if !condition() {
		return
	}
	return
}

func CacheEvict(f func() error, opts ...CacheOptions) {

}

func Any() bool {
	return true
}
