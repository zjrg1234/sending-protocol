package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"megin/library/logger"
	"megin/library/structs"
	"time"
)

// 基于go-redis v8封装
type RedisClient struct {
	rdb *redis.Client
	ctx context.Context
}

var redisClient = new(RedisClient)

func Connect(addr, password string) *RedisClient {
	if len(addr) <= 0 {
		addr = "127.0.0.1:6379"
	}

	redisClient.rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	redisClient.ctx = context.Background()

	err := Set("test", 1, time.Minute)
	if err != nil {
		logger.Fatal("Redis Error", zap.Error(err))
	}
	return redisClient
}

// 不过期
func SetForever(key string, value any) error {
	return redisClient.rdb.Set(redisClient.ctx, key, value, 0).Err()
}

// 设置过期时间,为了防把过期时间误传,如果不到1秒的,会当成秒处理
func Set(key string, value any, expiration time.Duration) error {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	return redisClient.rdb.Set(redisClient.ctx, key, value, expiration).Err()
}

func LPush(key string, value any, expiration time.Duration) error {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	err := redisClient.rdb.LPush(redisClient.ctx, key, value).Err()
	if err != nil {
		return err
	}
	err = redisClient.rdb.Expire(redisClient.ctx, key, expiration).Err()
	return err
}

func RangeAll(key string) ([]string, error) {
	cmd := redisClient.rdb.LRange(redisClient.ctx, key, 0, -1)
	var list []string
	if cmd.Err() != nil {
		return list, cmd.Err()
	}
	list = cmd.Val()
	return list, nil
}

func Lock(key string, value any, expiration time.Duration) bool {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	cmd := redisClient.rdb.SetNX(redisClient.ctx, key, value, expiration)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val()
}

// GET
func Get(key string) *redis.StringCmd {
	return redisClient.rdb.Get(redisClient.ctx, key)
}

//key:string,value:struct
func HMSet(key string, values any, expiration time.Duration) *redis.BoolCmd {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	return redisClient.rdb.HMSet(redisClient.ctx, key, values)
}

func HMGet(key string) *redis.StringStringMapCmd {
	return redisClient.rdb.HGetAll(redisClient.ctx, key)
}

//key:string,value:struct
func HSet(key string, field string, values any, expiration time.Duration) *redis.IntCmd {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	return redisClient.rdb.HSet(redisClient.ctx, key, field, values)
}

func HGetAll(key string) *redis.StringStringMapCmd {
	return redisClient.rdb.HGetAll(redisClient.ctx, key)
}

func HGet(key, field string) *redis.StringCmd {
	return redisClient.rdb.HGet(redisClient.ctx, key, field)
}

func Incr(key string, val float64) *redis.FloatCmd {
	return redisClient.rdb.IncrByFloat(redisClient.ctx, key, val)
}

func Delete(key string) *redis.IntCmd {
	return redisClient.rdb.Del(redisClient.ctx, key)
}

func SetStruct(key string, structVal any, expiration time.Duration) *redis.BoolCmd {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}

	rdb := redisClient.rdb
	mapVal := structs.ToMap(structVal)
	boolCmd := rdb.HMSet(redisClient.ctx, key, mapVal)
	redisClient.rdb.Expire(redisClient.ctx, key, expiration)
	return boolCmd
}

func GetStruct[T any](key string) (*T, error) {
	val := HMGet(key)
	t := new(T)
	//需要redis tag
	err := Scan(val, t, "json")
	return t, err
}

func Scan(cmd *redis.StringStringMapCmd, dest interface{}, tag ...string) error {
	if cmd.Err() != nil {
		return cmd.Err()
	}

	fieldTag := "json"
	if tag != nil && len(tag) > 0 {
		fieldTag = tag[0]
	}
	strct, err := structs.Struct(dest, fieldTag)
	if err != nil {
		return err
	}

	for k, v := range cmd.Val() {
		_, err := strct.Scan(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
