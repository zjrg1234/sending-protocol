package redis

import (
	"context"
	"encoding/json"
	"fmt"
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
type ClientInfo struct {
	TransmitterId       string `json:"transmitter_id"`
	ReceiverId          string `json:"receiver_id"`
	ReceiverHostPort    string `json:"receiver_host_port"`
	TransmitterHostPort string `json:"transmitter_host_port"`
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

// 存储链接的客户端信息
func SaveClientInfo(clientKey string, info *ClientInfo) error {
	key := clientKey

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	// 设置过期时间10秒钟
	return redisClient.rdb.Set(redisClient.ctx, key, data, 0).Err()
}

// 获取客户端信息 //receiveid 或者 transmitterId
func GetClientInfo(clientID string) (*ClientInfo, error) {
	key := clientID

	data, err := redisClient.rdb.Get(redisClient.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var info ClientInfo
	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// 删除客户端信息
func (r *RedisClient) DeleteClientInfo(clientID string) error {
	key := fmt.Sprintf("clientTransmitterId:%s", clientID)
	return r.rdb.Del(redisClient.ctx, key).Err()
}

//// 清理过期客户端
//func (r *RedisClient) CleanupExpiredClients() {
//	clients, err := r.GetAllClients()
//	if err != nil {
//		return
//	}
//
//	now := time.Now().Unix()
//	for _, client := range clients {
//		if now-client.LastSeen > 120 { // 2分钟未活跃
//			r.DeleteClientInfo(client.TransmitterId)
//			log.Printf("清理过期客户端: %s", client.TransmitterId)
//		}
//	}
//}

//// 获取所有在线客户端 先不实现
//func (r *RedisClient) GetAllClients() ([]*ClientInfo, error) {
//	var clients []*ClientInfo
//
//	keys, err := r.rdb.Keys(r.ctx, "client:*").Result()
//	if err != nil {
//		return nil, err
//	}
//
//	for _, key := range keys {
//		data, err := r.rdb.Get(r.ctx, key).Result()
//		if err != nil {
//			continue
//		}
//
//		var client ClientInfo
//		if json.Unmarshal([]byte(data), &client) == nil {
//			clients = append(clients, &client)
//		}
//	}
//
//	return clients, nil
//}

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
