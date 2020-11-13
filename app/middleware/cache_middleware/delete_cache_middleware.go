package cache_middleware

import (
	"errors"

	"com.github.gin-common/common/caches"
	"com.github.gin-common/common/controllers"
	"github.com/gin-gonic/gin"
)

type cacheKeysOption []string

func (c cacheKeysOption) Apply(context *gin.Context) {
	context.Set("cacheNames", c)
}

func WithCacheNames(cacheNames []string) controllers.MiddlewareOption {
	return cacheKeysOption(cacheNames)
}

type DeleteCacheMiddleware struct {
	cache caches.CacheProvider
}

func (middleware *DeleteCacheMiddleware) Init(cache caches.CacheProvider) {
	middleware.cache = cache
}

func (middleware *DeleteCacheMiddleware) Before(context *gin.Context) (err error) {
	return
}

func (middleware *DeleteCacheMiddleware) getCacheKeys(context *gin.Context) ([]string, error) {
	cacheName, exists := context.Get("cacheNames")
	if exists {
		if s, ok := cacheName.([]string); ok {
			return s, nil
		}
	}
	return nil, errors.New("must provide cache name")
}

func (middleware *DeleteCacheMiddleware) After(context *gin.Context) (err error) {
	// 删除缓存
	var cacheKeys []string
	cacheKeys, err = middleware.getCacheKeys(context)
	err = middleware.cache.DeleteMulti(cacheKeys...)

	if err != nil {
		return
	}
	return
}

func (middleware *DeleteCacheMiddleware) DeniedBeforeAbortContext() bool {
	return false
}

func (middleware *DeleteCacheMiddleware) AllowAfterAbortContext() bool {
	return false
}
