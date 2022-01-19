package main

//https://www.cnblogs.com/liyutian/p/10050320.html
//https://blog.csdn.net/zjcjava/article/details/112817499?spm=1001.2101.3001.6650.3&utm_medium=distribute.pc_relevant.none-task-blog-2%7Edefault%7ECTRLIST%7Edefault-3.pc_relevant_aa&depth_1-utm_source=distribute.pc_relevant.none-task-blog-2%7Edefault%7ECTRLIST%7Edefault-3.pc_relevant_aa&utm_relevant_index=6
//https://mp.weixin.qq.com/s?__biz=Mzg2ODU1MTI0OA==&mid=2247484739&idx=1&sn=6fb754a4e88a04d9c14af17379be7eb2&chksm=ceabda7cf9dc536a7c0d84a474f8a79c1a7168c5f456c37133aa236d081f9cf785f3fbddc989&mpshare=1&scene=23&srcid=0115aKY9n5RCFt8bXp9OgiaQ&sharer_sharetime=1642248824024&sharer_shareid=e310c4d733aaa12c88d17d19e93aab6e#rd
//https://blog.csdn.net/zhougubei/article/details/120909312
import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/nacos-group/nacos-sdk-go/inner/uuid"
)

type RedisLock struct {
	redisdb redis.Client

	uid         string
	expire_time int
	// 加锁资源
	key string
	// 订阅频道名
	subscribe_name string
	// 加锁命令
	lock_lua string
	// 释放锁命令
	unlock_lua string
}

// 构造函数
func newRedisLock(redisdb redis.Client, key string, expire int) *RedisLock {
	return &RedisLock{
		redisdb:        redisdb,
		uid:            uuid.Must(uuid.NewV4()).String(),
		expire_time:    expire,
		key:            key,
		subscribe_name: "redisson_lock__channel:" + key,
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

		// KEYS[1]=order KEYS[2]=redisson_lock__channel:order  ARGV[1]=0/1 ARGV[2]=15(生存时间)  ARGV[3]=uid:thread1
		unlock_lua: `
		if (redis.call('exists', KEYS[1]) == 0) then 
			redis.call('publish', KEYS[2], ARGV[1])
			return 1
		end
		if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then 
			return nil
		end
		local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1)
		if (counter > 0 ) then
			redis.call('pexpire', KEYS[1], ARGV[2])
			return 0
		else
			redis.call('del', KEYS[1])
			redis.call('publish', KEYS[2], ARGV[1])
			return 1
		end; 
		return nil`,
	}
}

// 尝试加锁函数
func (lock *RedisLock) try_lock() bool {
	lock_script := redis.NewScript(lock.lock_lua)
	if lock_script == nil {
		fmt.Println("lock_script :", lock_script)
		return false
	}
	expire_time := strconv.Itoa(lock.expire_time * 1000)
	result, err := lock_script.Run(&lock.redisdb, []string{lock.key}, []string{lock.uid, expire_time}).Result()
	if result != nil {
		fmt.Println("lock result error:", err)
		return false
	}
	fmt.Println("lock result success:", err)
	return true
}

// 外部调用加锁函数
func (lock *RedisLock) Lock() {
	for {
		//尝试加锁
		result_lock := lock.try_lock()
		if result_lock {
			fmt.Print("lock try_lock")
			return
		}
		// 监听锁释放释放
		pubsub := lock.redisdb.Subscribe(lock.subscribe_name)
		for msg := range pubsub.Channel() {
			fmt.Printf("channel=%s message=%s", msg.Channel, msg.Payload)
			if msg.Payload == "0" {
				break
			}
		}
	}
}

// 释放锁
func (lock *RedisLock) Unlock() {
	unlock_script := redis.NewScript(lock.unlock_lua)
	if unlock_script == nil {
		fmt.Println("unlock_script :", unlock_script)
		return
	}
	//order KEYS[2]=redisson_lock__channel:order  ARGV[1]=0/1 ARGV[2]=15(生存时间)  ARGV[3]=uid:thread1
	expire_time := strconv.Itoa(lock.expire_time * 1000)
	result, err := unlock_script.Run(&lock.redisdb, []string{lock.key, lock.subscribe_name}, []string{"1", expire_time, lock.uid}).Result()
	if result != nil {
		fmt.Println("unlock result: ", result, err)
		return
	}
	fmt.Println("unlock result: ", result, err)
}

func (lock *RedisLock) ShowLockInfo() {
	data, err := lock.redisdb.HGetAll(lock.key).Result()
	fmt.Println("lock info: ", data, err)
}

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

	redis_lock1 := newRedisLock(*redis_client, "order", 30)
	redis_lock2 := newRedisLock(*redis_client, "order", 30)

	redis_lock1.Lock()
	fmt.Println("redis_lock1 lock 1 success")
	redis_lock1.ShowLockInfo()

	time.Sleep(time.Duration(10) * time.Second)
	redis_lock1.Lock()
	fmt.Println("redis_lock1 lock 2 success")
	redis_lock1.ShowLockInfo()

	redis_lock1.Unlock()
	fmt.Println("redis_lock1 unlock 1 success")
	redis_lock1.ShowLockInfo()

	redis_lock1.Unlock()
	fmt.Println("redis_lock1 unlock 2 success")
	redis_lock1.ShowLockInfo()

	redis_lock2.Lock()
	fmt.Println("redis_lock2 lock success")
	redis_lock2.ShowLockInfo()

	redis_lock2.Unlock()
	fmt.Println("redis_lock2 unlock success")
	redis_lock2.ShowLockInfo()

	// lock1 加锁不释放
	// redis_lock1.Lock()
	// fmt.Println("redis_lock1 lock 2 success")
	// redis_lock1.ShowLockInfo()

	// redis_lock2.Lock()
	// fmt.Println("redis_lock1 lock 2 success")
	// redis_lock1.ShowLockInfo()

}
