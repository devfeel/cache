package main

import (
	"fmt"
	"github.com/devfeel/cache"
	"time"
)

func main() {
	c := cache.GetRuntimeCache()
	c.Set("1", 1, 100)
	fmt.Println(cache.Must(c.GetString("1")))

	//创建一个新的内存缓存，与之前GetRuntimeCache的不相关
	c2 := cache.NewRuntimeCache()
	fmt.Println(c2.GetString("1"))

	redisServer := "redis://192.168.8.175:6329/0"
	readOnlyServer := "redis://192.168.8.175:6339/0"
	backupRedisServer := "redis://192.168.8.175:6379/0"
	redisCache := cache.GetRedisCache(redisServer)
	redisCache.SetBackupServer(backupRedisServer, 10, 10)

	err := redisCache.HSet("hsettest", "keytest", "1")
	if err != nil {
		fmt.Println(`redisCache.HSet("hsettest", "keytest", "1")`, err)
	}

	_, err = redisCache.Expire("hsettest", 60000)
	if err != nil{
		fmt.Println(`redisCache.Expire("hsettest", 60)`, err)
	}
	fmt.Println(redisCache.HGet("hsettest", "keytest"))
	cr := cache.GetRedisCachePoolConf(redisServer, 10, 100)
	cr.SetReadOnlyServer(readOnlyServer, 10, 10)
	cr.SetBackupServer(backupRedisServer, 10, 10)
	err = cr.Set("1", 1, 100000)
	if err != nil{
		fmt.Println(`Set("1", 1, 100)`, err)
	}
	for i:=0;i< 20;i++ {
		fmt.Println(cr.GetString("1"))
	}



	time.Sleep(time.Hour)
}
