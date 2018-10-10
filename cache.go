package cache

import (
	"github.com/devfeel/cache/redis"
	"github.com/devfeel/cache/runtime"
	"sync"
)

const (
	CacheType_Runtime = "runtime"
	CacheType_Redis   = "redis"
)

const (
	RedisConnPool_MaxIdle   = 5
	RedisConnPool_MaxActive = 20
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
		// Expire Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
		// timeout time duration is second
		Expire(key string, timeOutSeconds int) (int, error)
	}

	RedisCache interface {
		Cache
		// SetReadOnlyServer set readonly redis server
		SetReadOnlyServer(serverUrl string, maxIdle int, maxActive int)
		// SetBackupServer set backup redis server, only use to read
		SetBackupServer(serverUrl string, maxIdle int, maxActive int)

		/*---------- Hash -----------*/
		// HGet Returns the value associated with field in the hash stored at key.
		HGet(hashID string, field string) (string, error)
		// HMGet Returns the values associated with the specified fields in the hash stored at key.
		HMGet(hashID string, field ...interface{}) ([]string, error)
		// HSet Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
		// If field already exists in the hash, it is overwritten.
		HSet(hashID string, field string, val string) error
		// HGetAll Returns all fields and values of the hash stored at key
		HGetAll(hashID string) (map[string]string, error)
		// HSetNX Sets field in the hash stored at key to value, only if field does not yet exist
		HSetNX(hashID string, field string, val string) (string, error)
		// HDel Removes the specified fields from the hash stored at key.
		HDel(hashID string, fields ...interface{}) (int, error)
		// HExists Returns if field is an existing field in the hash stored at key
		HExists(hashID string, field string) (int, error)
		// HIncrBy Increments the number stored at field in the hash stored at key by increment.
		HIncrBy(hashID string, field string, increment int) (int, error)
		// HIncrByFloat Increment the specified field of a hash stored at key, and representing a floating point number, by the specified increment
		HIncrByFloat(hashID string, field string, increment float64) (float64, error)
		// HKeys Returns all field names in the hash stored at key.
		HKeys(hashID string) ([]string, error)
		// HLen Returns the number of fields contained in the hash stored at key
		HLen(hashID string) (int, error)
		// HVals Returns all values in the hash stored at key
		HVals(hashID string) ([]string, error)
		// GetJsonObj get obj with SetJsonObj key
		GetJsonObj(key string, result interface{}) error
		// SetJsonObj set obj use json encode string
		SetJsonObj(key string, val interface{}) (interface{}, error)

		/*---------- List -----------*/
		// BLPop BLPOP is a blocking list pop primitive.
		// It is the blocking version of LPOP because it blocks the connection when there are no elements to pop from any of the given lists
		BLPop(key ...interface{}) (map[string]string, error)
		// BRPOP is a blocking list pop primitive
		// It is the blocking version of RPOP because it blocks the connection when there are no elements to pop from any of the given lists
		BRPop(key ...interface{}) (map[string]string, error)
		// BRPOPLPUSH is a operation like RPOPLPUSH but blocking
		BRPopLPush(source string, destination string) (string, error)
		// LIndex return element which subscript is index,
		// if index is -1, return last one element of list and so on
		LIndex(key string, index int) (string, error)
		// LInsert Inserts value in the list stored at key either before or after the reference value pivot.
		LInsert(key string, direction string, pivot string, value string) (int, error)
		// LLen return length of list
		LLen(key string) (int, error)
		// LPop remove and return head element of list
		LPop(key string) (string, error)
		// LPush Insert all the specified values at the head of the list stored at key
		LPush(key string, value ...interface{}) (int, error)
		// LPushX insert an element at the head of the list
		LPushX(key string, value string) (int, error)
		// LRange Returns the specified elements of the list stored at key
		LRange(key string, start int, end int) ([]string, error)
		// LRem Removes the first count occurrences of elements equal to value from the list stored at key.
		LRem(key string, count int, value string) (int, error)
		// LSet Sets the list element at index to value
		LSet(key string, index int, value string) (string, error)
		// LTrim Trim an existing list so that it will contain only the specified range of elements specified
		LTrim(key string, start int, stop int) (string, error)
		// RPop Removes and returns the last element of the list stored at key
		RPop(key string) (string, error)
		// RPopLPush Atomically returns and removes the last element (tail) of the list stored at source, and pushes the element at the first element (head) of the list stored at destination
		RPopLPush(source string, destination string) (string, error)
		// RPush Insert all the specified values at the tail of the list stored at key.
		RPush(key string, value ...interface{}) (int, error)
		// RPushX Inserts value at the tail of the list stored at key, only if key already exists and holds a list
		RPushX(key string, value ...interface{}) (int, error)

		/*---------- Set -----------*/
		// SAdd Add the specified members to the set stored at key
		SAdd(key string, value ...interface{}) (int, error)
		// SCard Returns the set cardinality (number of elements) of the set stored at key
		SCard(key string) (int, error)
		// SDiff Returns the members of the set resulting from the difference between the first set and all the successive sets
		SDiff(key ...interface{}) ([]string, error)
		// SDiffStore This command is equal to SDIFF, but instead of returning the resulting set, it is stored in destination
		SDiffStore(destination string, key ...interface{}) (int, error)
		// SInter Returns the members of the set resulting from the intersection of all the given sets.
		SInter(key ...interface{}) ([]string, error)
		// SInterStore This command is equal to SINTER, but instead of returning the resulting set, it is stored in destination
		SInterStore(destination string, key ...interface{}) (int, error)
		// SIsMember Returns if member is a member of the set stored at key.
		SIsMember(key string, value string) (bool, error)
		// SMembers Returns all the members of the set value stored at key.
		SMembers(key string) ([]string, error)
		// SMove Move member from the set at source to the set at destination
		SMove(source string, destination string, value string) (bool, error)
		// SPop Removes and returns one or more random elements from the set value store at key.
		SPop(key string) (string, error)
		// SRandMember When called with just the key argument, return a random element from the set value stored at key
		SRandMember(key string, count int) ([]string, error)
		// SRem Remove the specified members from the set stored at key
		SRem(key string, value ...interface{}) (int, error)
		// SUnion Returns the members of the set resulting from the union of all the given sets
		SUnion(key ...interface{}) ([]string, error)
		// SUnionStore This command is equal to SUNION, but instead of returning the resulting set, it is stored in destination
		SUnionStore(destination string, key ...interface{}) (int, error)

		/*---------- sorted set -----------*/
		// ZAdd Adds all the specified members with the specified scores to the sorted set stored at key
		ZAdd(key string, score int64, member interface{}) (int, error)
		// ZCount Returns the number of elements in the sorted set at key with a score between min and max
		ZCount(key string, min, max int64) (int, error)
		// ZRem Removes the specified members from the sorted set stored at key. Non existing members are ignored.
		ZRem(key string, member ...interface{}) (int, error)
		// ZCard Returns the sorted set cardinality (number of elements) of the sorted set stored at key.
		ZCard(key string) (int, error)
		// ZRank Returns the rank of member in the sorted set stored at key, with the scores ordered from low to high
		ZRank(key, member string) (int, error)
		// ZRange Returns the specified range of elements in the sorted set stored at key
		ZRange(key string, start, stop int64) ([]string, error)
		// ZRangeByScore Returns all the elements in the sorted set at key with a score between min and max (including elements with score equal to min or max).
		ZRangeByScore(key string, start, stop string, isWithScores bool) ([]string, error)
		// ZREVRangeByScore Returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min). In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
		ZREVRangeByScore(key string, max, min string, isWithScores bool) ([]string, error)
		// ZRange Returns the specified range of elements in the sorted set stored at key
		ZRevRange(key string, start, stop int64) ([]string, error)

		//****************** PUB/SUB *********************
		// Publish Posts a message to the given channel.
		Publish(channel string, message interface{}) (int64, error)

		// Subscribe Subscribes the client to the specified channels
		Subscribe(receive chan redis.Message, channels ...interface{}) error

		//****************** lua scripts *********************
		// EVAL used to evaluate scripts using the Lua interpreter built into Redis starting from version 2.6.0
		EVAL(script string, argsNum int, arg ...interface{}) (interface{}, error)
	}
)

func Must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return i
}

//get cache by gived ctype
//if set CacheType_Redis, must set serverUrl
func GetCache(ctype string, serverUrl ...string) Cache {
	switch ctype {
	case CacheType_Runtime:
		return GetRuntimeCache()
	case CacheType_Redis:
		if len(serverUrl) <= 0 {
			panic("GetRedisCache lost serverip!")
		}
		return GetRedisCache(serverUrl[0])
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
//must set serverIp like "redis://:password@10.0.1.11:6379/0"
func GetRedisCache(serverUrl string) RedisCache {
	return GetRedisCachePoolConf(serverUrl, RedisConnPool_MaxIdle, RedisConnPool_MaxActive)
}

//get redis cache
//must set serverIp like "redis://:password@10.0.1.11:6379/0"
func GetRedisCachePoolConf(serverUrl string, maxIdle int, maxActive int) RedisCache {
	if maxIdle <= 0 {
		maxIdle = RedisConnPool_MaxIdle
	}
	if maxActive < 0 {
		maxActive = RedisConnPool_MaxActive
	}
	c, ok := redisCacheMap[serverUrl]
	if !ok {
		c = NewRedisCache(serverUrl, maxIdle, maxActive)
		redisCacheLock.Lock()
		redisCacheMap[serverUrl] = c
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
//must set serverIp like "redis://:password@10.0.1.11:6379/0"
func NewRedisCache(serverUrl string, maxIdle int, maxActive int) RedisCache {
	return redis.NewRedisCache(serverUrl, maxIdle, maxActive)
}
