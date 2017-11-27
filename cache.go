package cache

import (
	"github.com/devfeel/cache/redis"
	"github.com/devfeel/cache/runtime"
	"sync"
)

const (
	CacheType_Runtime = "runtime"
	CacheType_Redis   = "redis"
	CacheType_MemCache = "memcache"
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
		HSetNx(hashID string, field string, val string) (string, error)
		HDel(hashID string, fields ...interface{}) (int64, error)
		HExists(hashID string, field string) (int64, error)
		HIncrBy(hashID string, field string) (int64, error)
		HIncrByFloat(hashID string, field string) (string, error)
		HKeys(hashID string) ([]string, error)
		HLen(hashID string) (int64, error)
		HVals(hashID string) ([]string, error)

		BLPop(args ...string)([]string, error)
		//BRPOP is a blocking list pop primitive
		BRPop(key string) (string, error)
		//BRPOPLPUSH is a operation like RPOPLPUSH but blocking
		BRPopLPush(args ...string)([]string, error)
		//return element which subscript is index,
		// if index is -1, return last one element of list and so on
		LIndex(key string, index int64)(string, error)
		//LINSERT key BEFORE|AFTER pivot value
		LInsertBefore()()
		//insert a new element after one element of list
		LInsertAfter()()
		//return length of list
		LLen() (int64, error)
		//remove and return head element of list
		LPop()()
		//Insert all the specified values at the head of the list stored at key
		LPush(key string, val string) (int64, error)
		//insert an element at the head of the list
		LPushX()()
		//LRANGE key start stop
		LRange(key string, start int64, end int64)([]string, error)

		LRem(key string, count int64, value string)(int64, error)

		LSet(key string, index int64, value string)(string, error)

		LTrim(key string, start int64, stop int64)(string, error)

		RPop(key string)(string, error)

		RPopLPush(source string, destination string)(string, error)
		//RPUSH key value [value ...]
		RPush(key string, value ...string) int64
		//push a value to list only if list is exist and return length of list after push
		// or return 0
		RPushX(key string, value ...string) int64


		SAdd(key string, value ...string)(int64, error)
		SCard(key string) (int64, error)
		SDiff(key ...string)([]string, error)
		SDiffStore(key ...string)([]string, error)
		SInter(key ...string)([]string, error)
		SInterStore(destination string, key ...string)(int64, error)
		SIsMember(key string, value string)(int64, error)
		SMembers(key string)([]string, error)
		SMove(source string, destination string, value string)(int64, error)
		SPop(key string)(string, error)
		SRandMember(key string, count int64)([]string, error)
		SRem(key string, value ...string)(int64, error)
		SUnion(key ...string)([]string, error)
		SUnionStore(destination string, key ...string)(int64, error)




	}

	MemCache interface {
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

//new runtime cache
func NewRuntimeCache() Cache {
	return runtime.NewRuntimeCache()
}

//new redis cache
//must set serverIp like "10.0.1.11:6379"
func NewRedisCache(serverIp string) *RedisCache {
	return redis.NewRedisCache(serverIp)
}
