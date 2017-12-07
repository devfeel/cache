package cache

import (
	"github.com/devfeel/cache/redis"
	"github.com/devfeel/cache/runtime"
	"sync"
	"github.com/devfeel/cache/memcached"
)

const (
	CacheType_Runtime   = "runtime"
	CacheType_Redis     = "redis"
	CacheType_MemCached = "memcached"
)

var (
	runtime_cache  Cache
	redisCacheMap  map[string]RedisCache
	redisCacheLock *sync.RWMutex
)

func init() {
	redisCacheMap = make(map[string]RedisCache)
	redisCacheLock = new(sync.RWMutex)
}

type (
	Cache interface {
		// Exist return true if value cached by given key
		Exists(key string) (bool, error)
		// Get returns value by given key
		Get(key string) (interface{}, error)
		// GetString returns value string format by given key
		GetString(key string) (string, error)
		// GetInt returns value int format by given key
		GetInt(key string) (int, error)
		// GetInt64 returns value int64 format by given key
		GetInt64(key string) (int64, error)
		// Set cache value by given key
		Set(key string, v interface{}, ttl int64) error
		// Incr increases int64-type value by given key as a counter
		// if key not exist, before increase set value with zero
		Incr(key string) (int64, error)
		// Decr decreases int64-type value by given key as a counter
		// if key not exist, before increase set value with zero
		Decr(key string) (int64, error)
		// Delete delete cache item by given key
		Delete(key string) error
		// ClearAll clear all cache items
		ClearAll() error
	}

	RedisCache interface {
		Cache
		//Returns the value associated with field in the hash stored at key.
		HGet(hashID string, field string) (string, error)
		//Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
		//If field already exists in the hash, it is overwritten.
		HSet(hashID string, field string, val string) error

		HGetAll(hashID string) (map[string]string, error)
		HSetNX(hashID string, field string, val string) (string, error)
		HDel(hashID string, fields ...interface{}) (int, error)
		HExists(hashID string, field string) (int, error)
		HIncrBy(hashID string, field string, increment int) (int, error)
		HIncrByFloat(hashID string, field string, increment float64) (float64, error)
		HKeys(hashID string) ([]string, error)
		HLen(hashID string) (int, error)
		HVals(hashID string) ([]string, error)

		BLPop(key ...interface{}) (map[string]string, error)
		//BRPOP is a blocking list pop primitive
		BRPop(key ...interface{}) (map[string]string, error)
		//BRPOPLPUSH is a operation like RPOPLPUSH but blocking
		BRPopLPush(source string, destination string) (string, error)
		//return element which subscript is index,
		// if index is -1, return last one element of list and so on
		LIndex(key string, index int) (string, error)
		//LINSERT key BEFORE|AFTER pivot value
		LInsert(key string, direction string, pivot string, value string) (int, error)
		//return length of list
		LLen(key string) (int, error)
		//remove and return head element of list
		LPop(key string) (string, error)
		//Insert all the specified values at the head of the list stored at key
		LPush(key string, value ...interface{}) (int, error)
		//insert an element at the head of the list
		LPushX(key string, value string) (int, error)
		//LRANGE key start stop
		LRange(key string, start int, end int) ([]string, error)

		LRem(key string, count int, value string) (int, error)

		LSet(key string, index int, value string) (string, error)

		LTrim(key string, start int, stop int) (string, error)

		RPop(key string) (string, error)

		RPopLPush(source string, destination string) (string, error)
		//RPUSH key value [value ...]
		RPush(key string, value ...interface{}) (int, error)
		//push a value to list only if list is exist and return length of list after push
		// or return 0
		RPushX(key string, value ...interface{}) (int, error)

		SAdd(key string, value ...interface{}) (int, error)
		SCard(key string) (int, error)
		SDiff(key ...interface{}) ([]string, error)
		SDiffStore(destination string, key ...interface{}) (int, error)
		SInter(key ...interface{}) ([]string, error)
		SInterStore(destination string, key ...interface{}) (int, error)
		SIsMember(key string, value string) (bool, error)
		SMembers(key string) ([]string, error)
		SMove(source string, destination string, value string) (bool, error)
		SPop(key string) (string, error)
		SRandMember(key string, count int) ([]string, error)
		SRem(key string, value ...interface{}) (int, error)
		SUnion(key ...interface{}) ([]string, error)
		SUnionStore(destination string, key ...interface{}) (int, error)
	}

	//memcached interface
	MemcachedCache interface {
		Cache
	}
)

func Must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return i
}

//get cache by gived ctype
//if set CacheType_Redis, must set serverip
func GetCache(ctype string, serverip ...string) Cache {
	switch ctype {
	case CacheType_Runtime:
		return GetRuntimeCache()
	case CacheType_Redis:
		if len(serverip) <= 0 {
			panic("GetRedisCache lost serverip!")
		}
		return GetRedisCache(serverip[0])
	case CacheType_MemCached:
		if len(serverip) <= 0 {
			panic("GetMemcachedCache lost serverip!")
		}
		return GetMemcachedCache(serverip...)
	default:
		return GetRuntimeCache()
	}
}

//get runtime cache
func GetRuntimeCache() Cache {
	if runtime_cache == nil {
		runtime_cache = NewRuntimeCache()
	}
	return runtime_cache
}

//get redis cache
//must set serverIp like "10.0.1.11:6379"
func GetRedisCache(serverIp string) RedisCache {
	c, ok := redisCacheMap[serverIp]
	if !ok {
		c = NewRedisCache(serverIp)
		redisCacheLock.Lock()
		redisCacheMap[serverIp] = c
		redisCacheLock.Unlock()
		return c

	} else {
		return c
	}
}

//set server like "127.0.0.1:11211"
func GetMemcachedCache(server ...string) Cache{
	return NewMemcachedCache(server...)
}

//new runtime cache
func NewRuntimeCache() Cache {
	return runtime.NewRuntimeCache()
}

//new redis cache
//must set serverIp like "10.0.1.11:6379"
func NewRedisCache(serverIp string) RedisCache {
	return redis.NewRedisCache(serverIp)
}

func NewMemcachedCache(server ...string) MemcachedCache{
	return memcached.NewMemcachedCache(server...)
}

