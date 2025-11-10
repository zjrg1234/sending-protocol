package redis_test

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"megin/library/datetime"
	"megin/library/logger"
	redisClient "megin/library/redis"
	"megin/library/structs"
	"megin/system/config"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	ExampleClient()
}

func TestRedisClientList(t *testing.T) {
	redisClient.Connect("127.0.0.1:6379", "")
	//标准用法:
	err := redisClient.LPush("list1", "120", time.Second*3600)
	if err != nil {
		fmt.Println(err)
	}

	list, err := redisClient.RangeAll("list1")
	fmt.Println(list, err)
}

func TestRedisClientSet(t *testing.T) {
	redisClient.Connect("127.0.0.1:6379", "")
	//标准用法:
	err := redisClient.Set("aa", "120", time.Second*1)
	if err != nil {
		fmt.Println(err)
	}
	//等价于上面的方法过期时间也是1秒。防误传
	err = redisClient.Set("aa", "121", 1)
	//返回:121
	v2, er := redisClient.Get("aa").Int()
	fmt.Println(v2, er)

	//TODO::这个真坑爹
	//返回:"get aa: 121" 不是你以为的"121"
	v1 := redisClient.Get("aa").String()
	//应该当用如下方法
	//返回:"121"
	v11 := redisClient.Get("aa").Val()
	fmt.Println(v1, v11)
	//然而要是获取别的类型,就可以直接这么用

	val := redisClient.Get("aa").Val()
	fmt.Println("过期前Get:", val)

	//停留2秒,让这个值过期
	time.Sleep(time.Second * 2)
	val = redisClient.Get("aa").Val()
	fmt.Println("过期后Get", val)
}

// 弄了一天
func TestRedisClientHMSet(t *testing.T) {
	redisClient.Connect("127.0.0.1:6379", "")
	type TestRedis struct {
		AA   string          `json:"aa" redis:"aa"`
		BB   int             `json:"bb" redis:"bb"`
		Time datetime.DBTime `json:"time" redis:"time"`
	}
	//encoding.BinaryMarshaler()
	val := TestRedis{AA: "test1111", BB: 12, Time: datetime.Now()}
	m := structs.ToMap(val)
	setCmd := redisClient.HMSet("bbb", m, time.Second*10)
	test3 := redisClient.HMGet("bbb")
	var val2 = TestRedis{}
	err := redisClient.Scan(test3, &val2)
	formatted := val2.Time.String()
	fmt.Println(setCmd, test3, err, formatted)
}

func TestRedisClientHSet(t *testing.T) {
	redisClient.Connect("127.0.0.1:6379", "")
	setCmd := redisClient.HSet("hsetteset", "wwww", "vvvvvv", time.Second*100)
	test3 := redisClient.HGet("hsetteset", "wwww")
	all := redisClient.HGetAll("hsetteset")
	for key, val := range all.Val() {
		fmt.Println(key, val)
	}
	fmt.Println(setCmd, test3, all)
}

var ctx = context.Background()

func ExampleClient() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get(ctx, "key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}

func TestMain(m *testing.M) {
	fmt.Println("run main")
	//解析配置文件
	config.InitConfig("../../resources/config-dev.yaml")
	//Log初始化
	logger.InitLog(logger.LogConfig{LogInConsole: true})
	m.Run()
}
