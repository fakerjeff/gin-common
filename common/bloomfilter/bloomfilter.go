package bloomfilter

import (
	"context"
	"fmt"
	"strconv"

	"com.github.gin-common/util"

	"github.com/go-redis/redis/v8"
)

type BFOption struct {
	bloomFilterEnable bool
	bloomFilter       BloomFilter
	filterKey         string
}

func (opt *BFOption) Enable() bool {
	return opt.bloomFilterEnable
}

func (opt *BFOption) Filter() BloomFilter {
	return opt.bloomFilter
}

func (opt *BFOption) Key() string {
	return opt.filterKey
}

func (opt *BFOption) WithOption(options ...BFOptions) {
	for _, o := range options {
		o.apply(opt)
	}
}

type BFOptions interface {
	apply(opt *BFOption)
}

type BFEnableOption bool

func (b BFEnableOption) apply(opt *BFOption) {
	opt.bloomFilterEnable = bool(b)
}

type RedisBFOption RedisBloomFilter

func (b RedisBFOption) apply(opt *BFOption) {
	filter := RedisBloomFilter(b)
	opt.bloomFilter = &filter
}

type FilterKeyOption string

func (s FilterKeyOption) apply(opt *BFOption) {
	opt.filterKey = string(s)
}

type BloomFilter interface {
	Add(key string, item string) (bool, error)
	Exists(key string, item string) (bool, error)
	Info(key string) (map[string]int64, error)
	AddMulti(key string, items ...interface{}) ([]bool, error)
	ExistsMulti(key string, items ...interface{}) ([]bool, error)
}

type RedisBloomFilter struct {
	client *RedisBloomFilterClient
}

func (f *RedisBloomFilter) Init(client *RedisBloomFilterClient) {
	f.client = client
}

func (f *RedisBloomFilter) Add(key string, item string) (bool, error) {
	return f.client.Add(key, item)
}

func (f *RedisBloomFilter) Exists(key string, item string) (bool, error) {
	return f.client.Exists(key, item)
}

func (f *RedisBloomFilter) Info(key string) (map[string]int64, error) {
	return f.client.Info(key)
}

func (f *RedisBloomFilter) AddMulti(key string, items ...interface{}) ([]bool, error) {
	return f.client.MAdd(key, items...)
}

func (f *RedisBloomFilter) ExistsMulti(key string, items ...interface{}) ([]bool, error) {
	return f.client.MExists(key, items)
}

type RedisBloomFilterClient struct {
	// 基于redis的布隆过滤器实现(依赖RedisBloom)
	rdb *redis.Client
	ctx context.Context
}

func (c *RedisBloomFilterClient) Init(rdb *redis.Client, ctx context.Context) {
	c.rdb = rdb
	c.ctx = ctx
}

func (c *RedisBloomFilterClient) Reserve(key string, errorRate float64, capacity uint64) error {
	_, err := c.rdb.Do(c.ctx, "BF.RESERVE", key, strconv.FormatFloat(errorRate, 'g', 16, 64), capacity).Result()
	return err
}

func (c *RedisBloomFilterClient) Add(key string, item string) (bool, error) {
	cmd := redis.NewBoolCmd(c.ctx, "BF.ADD", key, item)
	_ = c.rdb.Process(c.ctx, cmd)
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val(), nil
}

func (c *RedisBloomFilterClient) Exists(key string, item string) (bool, error) {
	cmd := redis.NewBoolCmd(c.ctx, "BF.EXISTS", key, item)
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val(), nil
}

func (c *RedisBloomFilterClient) Info(key string) (map[string]int64, error) {
	cmd := redis.NewStringIntMapCmd(c.ctx, "BF.INFO", key)
	_ = c.rdb.Process(c.ctx, cmd)

	if err := cmd.Err(); err != nil {
		return nil, err
	}
	return cmd.Val(), nil

}

func (c *RedisBloomFilterClient) MAdd(key string, items ...interface{}) ([]bool, error) {
	var args []interface{}
	args = append(args, "BF.MADD")
	args = append(args, key)
	args = append(args, items...)

	cmd := redis.NewBoolSliceCmd(c.ctx, args...)
	_ = c.rdb.Process(c.ctx, cmd)

	if err := cmd.Err(); err != nil {
		return nil, err
	}
	return cmd.Val(), nil
}

func (c *RedisBloomFilterClient) MExists(key string, items ...interface{}) ([]bool, error) {
	var args []interface{}
	args = append(args, "BF.MEXISTS")
	args = append(args, key)
	args = append(args, items...)

	cmd := redis.NewBoolSliceCmd(c.ctx, args...)
	_ = c.rdb.Process(c.ctx, cmd)

	if err := cmd.Err(); err != nil {
		return nil, err
	}

	return cmd.Val(), nil
}

func (c *RedisBloomFilterClient) Insert(key string, cap int64, errorRate float64, expansion int64, noCreate bool, nonScaling bool, items ...interface{}) ([]bool, error) {
	var args []interface{}
	args = append(args, "BF.INSERT")
	args = append(args, key)
	if cap > 0 {
		args = append(args, fmt.Sprintf("CAPACITY %d", cap))
	}

	if errorRate > 0 {
		args = append(args, fmt.Sprintf("ERROR %f", errorRate))
	}

	if expansion > 0 {
		args = append(args, fmt.Sprintf("EXPANSION %d", expansion))
	}
	if noCreate {
		args = append(args, "NOCREATE")
	}
	if nonScaling {
		args = append(args, "NONSCALING")
	}

	args = append(args, items...)
	cmd := redis.NewSliceCmd(c.ctx, args...)

	_ = c.rdb.Process(c.ctx, cmd)

	if err := cmd.Err(); err != nil {
		return nil, err
	}
	var innerRes bool
	var err error
	var res []bool

	for _, arrayPos := range cmd.Val() {
		innerRes, err = util.IsBool(arrayPos)
		if err == nil {
			res = append(res, innerRes)
		} else {
			break
		}
	}

	return res, nil
}

func (c *RedisBloomFilterClient) ScanDump(key string, iter int64) (int64, []byte, error) {
	cmd := redis.NewSliceCmd(c.ctx, "BF.SCANDUMP", key, iter)
	_ = c.rdb.Process(c.ctx, cmd)
	val, err := cmd.Result()
	if err != nil || len(val) != 2 {
		return 0, nil, err
	}
	var ok bool
	iter, ok = val[0].(int64)
	if !ok {
		return 0, nil, err
	}
	if val[1] == nil {
		return iter, nil, err
	}
	var data []byte
	data, ok = val[1].([]byte)
	if !ok {
		return 0, nil, err
	}
	return iter, data, err
}

func (c *RedisBloomFilterClient) LoadChunk(key string, iter int64, data []byte) (string, error) {
	cmd := redis.NewStringCmd(c.ctx, "BF.LOADCHUNK", key, iter, data)
	_ = c.rdb.Process(c.ctx, cmd)
	if err := cmd.Err(); err != nil {
		return "", err
	}
	return cmd.Val(), nil
}
