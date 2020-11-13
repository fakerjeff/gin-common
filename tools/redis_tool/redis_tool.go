package redis_tool

import (
	"fmt"
	"strconv"
	"sync"

	"com.github.gin-common/util"

	"github.com/go-redis/redis/v8"
)

var ginServerRdb *redis.Client
var ginServerRedisOnce sync.Once

func GetGinServerRdb() *redis.Client {
	ginServerRedisOnce.Do(func() {
		db, err := strconv.Atoi(util.GetDefaultEnv("REDIS_DB", "0"))
		util.PanicError(err)
		var maxRetries, poolSize, minIdleConns int
		maxRetries, err = strconv.Atoi(util.GetDefaultEnv("REDIS_MAX_RETRIES", "3"))
		util.PanicError(err)
		poolSize, err = strconv.Atoi(util.GetDefaultEnv("REDIS_POOL_SIZE", "10"))
		util.PanicError(err)
		minIdleConns, err = strconv.Atoi(util.GetDefaultEnv("REDIS_MIN_IDLE", "0"))

		ginServerRdb = redisConn(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", util.GetDefaultEnv("REDIS_HOST", "127.0.0.1"), util.GetDefaultEnv("REDIS_PORT", "6379")),
			Password:     util.GetDefaultEnv("REDIS_PASSWORD", ""),
			DB:           db,
			MaxRetries:   maxRetries,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
		})
		util.PanicError(err)
	})
	return ginServerRdb
}

func redisConn(opt *redis.Options) *redis.Client {
	var rdb *redis.Client
	rdb = redis.NewClient(opt)
	return rdb
}
