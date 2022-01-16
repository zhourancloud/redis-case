package main

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/nacos-group/nacos-sdk-go/inner/uuid"
)

var lock_command string = `-- 如果锁不存在，则通过hset设置它的值，并设置过期时间, KEYS[1] key值， ARGV
if (redis.call('exists', KEYS[1]) == 0) then 
	redis.call('hset', KEYS[1], ARGV[2], 1); 
	redis.call('pexpire', KEYS[1], ARGV[1]); 
	return nil
end
--如果锁已存在，并且锁的是当前线程，则通过hincrby给数值递增1
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then 
	redis.call('hincrby', KEYS[1], ARGV[2], 1)
	redis.call('pexpire', KEYS[1], ARGV[1])
	return nil
end
-- 如果锁已存在，但并非本线程，则返回过期时间ttl
	return redis.call('pttl', KEYS[1]);`

//https://blog.csdn.net/zjcjava/article/details/112817499?spm=1001.2101.3001.6650.3&utm_medium=distribute.pc_relevant.none-task-blog-2%7Edefault%7ECTRLIST%7Edefault-3.pc_relevant_aa&depth_1-utm_source=distribute.pc_relevant.none-task-blog-2%7Edefault%7ECTRLIST%7Edefault-3.pc_relevant_aa&utm_relevant_index=6
//https://mp.weixin.qq.com/s?__biz=Mzg2ODU1MTI0OA==&mid=2247484739&idx=1&sn=6fb754a4e88a04d9c14af17379be7eb2&chksm=ceabda7cf9dc536a7c0d84a474f8a79c1a7168c5f456c37133aa236d081f9cf785f3fbddc989&mpshare=1&scene=23&srcid=0115aKY9n5RCFt8bXp9OgiaQ&sharer_sharetime=1642248824024&sharer_shareid=e310c4d733aaa12c88d17d19e93aab6e#rd

func acquire_lock(redisdb *redis.Client, key string, expire int) bool {
	// lock_command := `if redis.call("GET", KEYS[1]) == ARGV[1] then
	// redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
	// 	return "OK"
	// else
	// 	return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
	// end`
	lock_command := `
	if (redis.call('exists', KEYS[1]) == 0) then 
		redis.call('hset', KEYS[1], ARGV[1], 1); 
		--redis.call('pexpire', KEYS[1], ARGV[2]); 
		return nil
	end
	if (redis.call('hexists', KEYS[1], ARGV[1]) == 1) then 
		redis.call('hincrby', KEYS[1], ARGV[1], 1)
		--redis.call('pexpire', KEYS[1], ARGV[2])
		return "OK1"
	end	
	return "OKen" .. KEYS[1] .. ' ' .. ARGV[1]
	`

	lock_script := redis.NewScript(lock_command)
	if lock_script == nil {
		fmt.Println("lock_script :", lock_script)
		return false
	}

	ids := uuid.Must(uuid.NewV4()).String()
	expire_time := strconv.Itoa(expire * 1000)
	n, err := lock_script.Run(redisdb, []string{key}, []string{ids, expire_time}).Result()
	fmt.Println("result: ", n, err)

	return true
}

func main() {
	redis_client := redis.NewClient(&redis.Options{
		Addr:     "120.78.127.165:6379",
		Password: "123456",
		DB:       0,
	})
	pong, err := redis_client.Ping().Result()
	// Output: PONG <nil>
	if err != nil {
		fmt.Println("pong: ", pong, "err: ", err)
		return
	}
	// retAll, err := redis_client.HGetAll("order").Result()
	// fmt.Println("keys *:", retAll, err)

	result := acquire_lock(redis_client, "order", 60)
	if result == false {
		fmt.Println("lock error")
		return
	}
	result = acquire_lock(redis_client, "order", 60)
	if result == false {
		fmt.Println("lock error")
		return
	}
	fmt.Println("lock success")
}
