package main

//https://www.cnblogs.com/liyutian/p/10050320.html
import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/nacos-group/nacos-sdk-go/inner/uuid"
)

type RedisLock struct {
	uid      string
	redisdb  redis.Client
	lock_lua string
}

// 构造函数
func newRedisLock(redisdb redis.Client) *RedisLock {
	return &RedisLock{
		uid:     uuid.Must(uuid.NewV4()).String(),
		redisdb: redisdb,

		lock_lua: `
		if (redis.call('exists', KEYS[1]) == 0) then 
			redis.call('hset', KEYS[1], ARGV[1], 1); 
			redis.call('pexpire', KEYS[1], ARGV[2]); 
			return nil
		end
		if (redis.call('hexists', KEYS[1], ARGV[1]) == 1) then 
			redis.call('hincrby', KEYS[1], ARGV[1], 1)
			redis.call('pexpire', KEYS[1], ARGV[2])
			return nil
		end	
		return redis.call('pttl', KEYS[1]);`,
	}
}

// 加锁函数
func (lock *RedisLock) AcquireLock(key string, expire int) bool {
	lock_script := redis.NewScript(lock.lock_lua)
	if lock_script == nil {
		fmt.Println("lock_script :", lock_script)
		return false
	}
	expire_time := strconv.Itoa(expire * 1000)
	n, err := lock_script.Run(&lock.redisdb, []string{key}, []string{lock.uid, expire_time}).Result()
	fmt.Println("result: ", n, err)

	return true
}

//解锁函数

func main() {
	redis_client := redis.NewClient(&redis.Options{
		Addr:     "120.78.127.165:6379",
		Password: "123456",
		DB:       0,
	})
	pong, err := redis_client.Ping().Result()
	if err != nil {
		fmt.Println("pong: ", pong, "err: ", err)
		return
	}

	redis_lock1 := newRedisLock(*redis_client)
	result := redis_lock1.AcquireLock("order", 15)
	fmt.Print("result: ", result)
	result = redis_lock1.AcquireLock("order", 15)
	fmt.Print("result: ", result)
}
