package cache_middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"com.github.gin-common/util"

	"com.github.gin-common/common/controllers"

	"com.github.gin-common/common/caches"
	"github.com/gin-gonic/gin"
)

type cacheResponseWriter struct {
	gin.ResponseWriter
	respBuffer *bytes.Buffer
}

func (w cacheResponseWriter) Write(data []byte) (int, error) {
	writer := io.MultiWriter(w, w.respBuffer)
	return writer.Write(data)
}

type AddCacheMiddleware struct {
	cache      caches.CacheProvider
	cached     bool
	respBuffer *bytes.Buffer
}

type expiresOption time.Duration

func (e expiresOption) Apply(ctx *gin.Context) {
	ctx.Set("expires", time.Duration(e))
}

func WithExpires(expires time.Duration) controllers.MiddlewareOption {
	return expiresOption(expires)
}

type cacheNameOption string

func (c cacheNameOption) Apply(ctx *gin.Context) {
	ctx.Set("cacheName", string(c))
}

func WithCacheName(cacheName string) controllers.MiddlewareOption {
	return cacheNameOption(cacheName)
}

func (middleware *AddCacheMiddleware) getCacheKey(context *gin.Context) (string, error) {
	// 根据上下文计算缓存key
	return "", nil
}

func (middleware *AddCacheMiddleware) Init(cache caches.CacheProvider) {
	middleware.cache = cache
}

func (middleware *AddCacheMiddleware) Before(context *gin.Context) (err error) {

	// 该中间件只针对于Get请求有效
	if context.Request.Method != http.MethodGet {
		return
	}

	var cacheKey string
	cacheKey, err = middleware.getCacheKey(context)
	if err != nil {
		return
	}

	var cacheValue string
	err = middleware.cache.Get(cacheKey, &cacheValue)
	if err != nil {
		return
	}

	if cacheValue != "" {
		middleware.cached = true
		_, err = context.Writer.Write([]byte(cacheValue))
		if err != nil {
			return
		}
		context.Abort()
	} else {
		context.Writer = cacheResponseWriter{
			ResponseWriter: context.Writer,
			respBuffer:     middleware.respBuffer,
		}
	}

	return
}

func (middleware *AddCacheMiddleware) getContextExpires(context *gin.Context) (time.Duration, error) {
	expires, exists := context.Get("expires")
	if exists {
		if t, ok := expires.(time.Duration); ok {
			return t, nil
		}
	}
	strExpires := util.GetDefaultEnv("DEFAULT_CACHE_EXPIRE", "300")
	i, err := strconv.Atoi(strExpires)
	if err != nil {
		return time.Duration(nil), err
	}
	return time.Duration(i) * time.Second, nil
}

func (middleware *AddCacheMiddleware) getContextCacheName(context *gin.Context) (string, error) {
	cacheName, exists := context.Get("cacheName")
	if exists {
		if s, ok := cacheName.(string); ok {
			return s, nil
		}
	}
	return "", errors.New("must provide cache name")
}

func (middleware *AddCacheMiddleware) After(context *gin.Context) (err error) {

	// 该中间件只针对于Get请求有效
	if context.Request.Method != http.MethodGet {
		return
	}

	if !middleware.cached {
		var cacheKey, cacheName string
		cacheKey, err = middleware.getCacheKey(context)
		if err != nil {
			return
		}
		cacheName, err = middleware.getContextCacheName(context)
		if err != nil {
			return
		}
		expires, err := middleware.getContextExpires(context)
		if err != nil {
			return
		}
		err = middleware.cache.HSet(fmt.Sprintf("cache:%s", cacheName), cacheKey, middleware.respBuffer.String(), expires)
		if err != nil {
			return
		}
	}
	return
}

func (middleware *AddCacheMiddleware) DeniedBeforeAbortContext() bool {
	return true
}

func (middleware *AddCacheMiddleware) AllowAfterAbortContext() bool {
	return false
}
