package caches

import (
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

type CacheError string

func newCacheError(s string) CacheError {
	return CacheError(s)
}

func (e CacheError) Error() string {
	return string(e)
}

var (
	ErrCacheMiss    = newCacheError("cache: key not found")
	ErrNotStored    = newCacheError("cache: not stored")
	ErrInvalidValue = newCacheError("cache: invalid value")
)

type cacheOption struct {
	key               string               // 缓存key
	bloomFilterOption bloomfilter.BFOption // 布隆过滤器配置
	cacheProvider     CacheProvider        // 缓存Provider
	expires           time.Duration        // 缓存过期时间
	serializer        util.Serializer      //缓存值序列化器
	condition         func() bool          //条件函数
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

type ConditionFuncOption func() bool

func (c ConditionFuncOption) apply(cacheOption *cacheOption) {
	cacheOption.condition = c
}

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

func CacheEnable(process func() (interface{}, error), ptr interface{}, opts ...CacheOptions) (r interface{}, e error) {
	// 缓存装饰方法,在存在缓存时读取缓存，不存在缓存时，从原方法中获取结果，并且将结果存如缓存（支持开启布隆过滤器（防穿透）、开启同步更新缓存）
	// @args
	// process 被装饰的处理方法
	// condition 缓存准入条件，若返回为false，则不缓存
	// opts 缓存配置
	// @return
	// r 返回值
	// e 返回异常

	// 缓存准入条件,若不符合则直接返回原方法
	options := newCacheOption(opts...)
	if options.condition != nil && !options.condition() {
		r, e = process()
		return
	}
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
		// 如果布隆过滤器中不存在请求key，则直接返回异常（判定为异常请求）
		var exists bool
		exists, e = filter.Exists(options.bloomFilterOption.Key(), key)
		if e != nil {
			e = newCacheError(e.Error())
			return
		}
		if !exists {
			e = newCacheError("item is not found")
			return
		}
	}
	// 若布隆过滤器中发现缓存，则尝试从缓存中获取结果
	var result string
	e = options.cacheProvider.Get(key, &result)
	if e != nil {
		e = newCacheError(e.Error())
		return
	}
	//若未获取结果，则尝试使用原处理方法获取结果，并更新缓存
	if result == "" {
		// 如果开启了同步模式，则需要给更新缓存的操作加锁（针对缓存key）
		if options.sync {
			// Todo： 此过程(更新缓存)考虑加锁，因为在多goroutine或分布式环境下同一时间相同key可能出现多次计算（分布式锁）
			// Todo：（实现方式建议：zk或etcd）
			e = options.cacheProvider.Get(key, &result)
			if e != nil {
				e = newCacheError(e.Error())
				return
			}
			if result == "" {
				r, e = process()
				if e != nil {
					return
				}
				e = options.cacheProvider.Set(key, r, options.expires)
				if e != nil {
					e = newCacheError(e.Error())
					return
				}
			}
		} else { //	反之则不加锁，直接请求原处理方法并更新缓存
			r, e = process()
			if e != nil {
				return
			}
			e = options.cacheProvider.Set(key, r, options.expires)
			if e != nil {
				e = newCacheError(e.Error())
				return
			}
		}
	}
	// 此处需要将查出来的结果反序列化
	e = options.serializer.Deserialize([]byte(result), ptr)
	if e != nil {
		e = newCacheError(e.Error())
		return
	}
	r = ptr
	return
}

func CachePut(process func() (interface{}, error), opts ...CacheOptions) (r interface{}, e error) {
	// 缓存装饰方法,执行处理方法，并将处理的结果写入缓存中
	// @args
	// process 被装饰的处理方法
	// condition 缓存准入条件，若返回为false，则不缓存
	// opts 缓存配置
	// @return
	// r 返回值
	// e 返回异常
	r, e = process()
	options := newCacheOption(opts...)

	if options.condition != nil && !options.condition() {
		return
	}

	key := options.key
	// 若没配置缓存key则直接返回原函数结果
	if key == "" {
		return
	}
	// 添加处理函数的结果在缓存
	e = options.cacheProvider.Set(key, r, options.expires)
	if e != nil {
		e = newCacheError(e.Error())
		return
	}
	// 判断是否开启布隆过滤器(防穿透)
	// 若开启则先检验key是否存在于布隆过滤器中，若不存在则添加到布隆过滤器中
	if options.bloomFilterOption.Enable() {
		// 若开启了布隆过滤器,但并未指定布隆过滤器，则直接返回原函数结果
		filter := options.bloomFilterOption.Filter()
		if filter == nil {
			return
		}
		// 如果布隆过滤器中不存在请求key，则将key添加到布容过滤器中
		var exists bool
		exists, e = filter.Exists(options.bloomFilterOption.Key(), key)
		if e != nil {
			e = newCacheError(e.Error())
			return
		}
		if !exists {
			_, e = filter.Add(options.bloomFilterOption.Key(), key)
			if e != nil {
				e = newCacheError(e.Error())
				return
			}
		}
	}
	return
}

func CacheEvict(process func() (interface{}, error), cacheProvider CacheProvider, cacheKeys ...string) (r interface{}, e error) {
	// 缓存装饰方法,删除缓存
	// @args
	// process 被装饰的处理方法
	// opts 缓存配置
	// @return
	// r 返回值
	// e 返回异常
	r, e = process()
	if e != nil {
		return
	}

	e = cacheProvider.DeleteMulti(cacheKeys...)
	if e != nil {
		e = newCacheError(e.Error())
		return
	}
	return
}

func Any() bool {
	return true
}
