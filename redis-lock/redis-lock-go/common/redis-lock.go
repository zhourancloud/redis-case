package main

//https://www.cnblogs.com/liyutian/p/10050320.html
//https://blog.csdn.net/zjcjava/article/details/112817499?spm=1001.2101.3001.6650.3&utm_medium=distribute.pc_relevant.none-task-blog-2%7Edefault%7ECTRLIST%7Edefault-3.pc_relevant_aa&depth_1-utm_source=distribute.pc_relevant.none-task-blog-2%7Edefault%7ECTRLIST%7Edefault-3.pc_relevant_aa&utm_relevant_index=6
//https://mp.weixin.qq.com/s?__biz=Mzg2ODU1MTI0OA==&mid=2247484739&idx=1&sn=6fb754a4e88a04d9c14af17379be7eb2&chksm=ceabda7cf9dc536a7c0d84a474f8a79c1a7168c5f456c37133aa236d081f9cf785f3fbddc989&mpshare=1&scene=23&srcid=0115aKY9n5RCFt8bXp9OgiaQ&sharer_sharetime=1642248824024&sharer_shareid=e310c4d733aaa12c88d17d19e93aab6e#rd

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

// 尝试加锁函数
func (lock *RedisLock) try_lock(key string, expire int) bool {
	lock_script := redis.NewScript(lock.lock_lua)
	if lock_script == nil {
		fmt.Println("lock_script :", lock_script)
		return false
	}
	expire_time := strconv.Itoa(expire * 1000)
	n, err := lock_script.Run(&lock.redisdb, []string{key}, []string{lock.uid, expire_time}).Result()
	fmt.Println("lock result n:", n)
	fmt.Println("lock result err:", err)
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
	redis_lock2 := newRedisLock(*redis_client)

	redis_lock1.try_lock("order", 15)
	//fmt.Println("1111111122222")
	redis_lock1.try_lock("order", 15)
	//fmt.Println("11111111111")
	redis_lock2.try_lock("order", 15)

}
