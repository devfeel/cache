package main

import (
	"fmt"
	"github.com/devfeel/cache"
)

func main() {
	c := cache.GetRuntimeCache()
	c.Set("1", 1, 100)
	fmt.Println(cache.Must(c.GetString("1")))

	//创建一个新的内存缓存，与之前GetRuntimeCache的不相关
	c2 := cache.NewRuntimeCache()
	fmt.Println(c2.GetString("1"))

	redisServer := "redis://192.168.8.175:6379/0"
	cr := cache.GetRedisCachePoolConf(redisServer, 10, 100)
	cr.Set("1", 1, 100)
	fmt.Println(cache.Must(cr.GetString("1")))

	c0 := cache.GetCache(cache.CacheType_Redis, redisServer)
	c0.Delete("1")
	fmt.Println(c0.GetString("1"))
}
