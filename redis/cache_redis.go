package redis

import (
	"errors"
	"fmt"
	"github.com/devfeel/cache/internal" //internal目录 不允许其他包调用, commit时候改回来
	"github.com/devfeel/cache/internal/hystrix"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
)

var (
	ZeroInt64 int64 = 0
)

const (
	LInsert_Before    = "BEFORE"
	LInsert_After     = "AFTER"
	HystrixErrorCount = 50
)

// Message represents a message notification.
type Message struct {
	// The originating channel.
	Channel string

	// The message data.
	Data []byte
}

// RedisCache is redis cache adapter.
// it contains serverIp for redis conn.
type redisCache struct {
	hystrix hystrix.Hystrix

	serverUrl string //connection string, like "redis://:password@10.0.1.11:6379/0"
	// Maximum number of idle connections in the pool.
	maxIdle int
	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	maxActive int

	//use to readonly server
	readOnlyServerUrl string
	readOnlyMaxIdle   int
	readOnlyMaxActive int

	//use to backup server
	backupServerUrl string
	backupMaxIdle   int
	backupMaxActive int
}

// NewRedisCache returns a new *RedisCache.
func NewRedisCache(serverUrl string, maxIdle int, maxActive int) *redisCache {
	cache := redisCache{serverUrl: serverUrl, maxIdle: maxIdle, maxActive: maxActive}
	cache.hystrix = hystrix.NewHystrix(cache.checkRedisAlive, nil)
	cache.hystrix.SetMaxFailedNumber(HystrixErrorCount)
	cache.hystrix.Do()
	return &cache
}

// SetReadOnlyServer set readonly redis server
func (ca *redisCache) SetReadOnlyServer(serverUrl string, maxIdle int, maxActive int) {
	ca.readOnlyServerUrl = serverUrl
	ca.readOnlyMaxActive = maxActive
	ca.readOnlyMaxIdle = maxIdle
}

// SetBackupServer set backup redis server, only use to read
func (ca *redisCache) SetBackupServer(serverUrl string, maxIdle int, maxActive int) {
	ca.backupServerUrl = serverUrl
	ca.backupMaxActive = maxActive
	ca.backupMaxIdle = maxIdle
}

// Exists check item exist in redis cache.
func (ca *redisCache) Exists(key string) (bool, error) {
	client := ca.getReadRedisClient()
	exists, err := client.Exists(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.Exists(key)
	}
	return exists, err
}

// Incr increase int64 counter in redis cache.
func (ca *redisCache) Incr(key string) (int64, error) {
	client := ca.getDefaultRedis()
	val, err := client.INCR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Decr decrease counter in redis cache.
func (ca *redisCache) Decr(key string) (int64, error) {
	client := ca.getDefaultRedis()
	val, err := client.DECR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Get cache from redis cache.
// if non-existed or expired, return nil.
func (ca *redisCache) Get(key string) (interface{}, error) {
	client := ca.getReadRedisClient()
	reply, err := client.GetObj(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.GetObj(key)
	}
	return reply, err
}

// GetString returns value string format by given key
// if non-existed or expired, return "".
func (ca *redisCache) GetString(key string) (string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.Get(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.Get(key)
	}
	return reply, err
}

// GetInt returns value int format by given key
// if non-existed or expired, return nil.
func (ca *redisCache) GetInt(key string) (int, error) {
	v, err := ca.GetString(key)
	if err != nil || v == "" {
		return 0, err
	} else {
		i, e := strconv.Atoi(v)
		if e != nil {
			return 0, err
		} else {
			return i, nil
		}
	}
}

// GetInt64 returns value int64 format by given key
// if non-existed or expired, return nil.
func (ca *redisCache) GetInt64(key string) (int64, error) {
	v, err := ca.GetString(key)
	if err != nil || v == "" {
		return ZeroInt64, err
	} else {
		i, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			return ZeroInt64, err
		} else {
			return i, nil
		}
	}
}

// Set cache to redis.
// ttl is second, if ttl is 0, it will be forever.
func (ca *redisCache) Set(key string, value interface{}, ttl int64) error {
	var err error
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	if ttl > 0 {
		_, err = client.SetWithExpire(key, value, ttl)
	} else {
		_, err = client.Set(key, value)
	}
	return err
}

// Delete item in redis cacha.
// if not exists, we think it's success
func (ca *redisCache) Delete(key string) error {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	_, err := client.Del(key)
	return err
}

// Expire Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
func (ca *redisCache) Expire(key string, timeOutSeconds int) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.Expire(key, timeOutSeconds)
	return reply, err
}

// GetJsonObj get obj with SetJsonObj key
func (ca *redisCache) GetJsonObj(key string, result interface{}) error {
	client := ca.getReadRedisClient()
	err := client.GetJsonObj(key, result)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.GetJsonObj(key, result)
	}
	return err
}

// SetJsonObj set obj use json encode string
func (ca *redisCache) SetJsonObj(key string, val interface{}) (interface{}, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.SetJsonObj(key, val)
	return reply, err
}

/*---------- Hash -----------*/
// HGet Returns the value associated with field in the hash stored at key.
func (ca *redisCache) HGet(key, field string) (string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HGet(key, field)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HGet(key, field)
	}
	return reply, err
}

// HMGet Returns the values associated with the specified fields in the hash stored at key.
func (ca *redisCache) HMGet(hashID string, field ...interface{}) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HMGet(hashID, field...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HMGet(hashID, field...)
	}
	return reply, err
}

// HGetAll Returns all fields and values of the hash stored at key
func (ca *redisCache) HGetAll(key string) (map[string]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HGetAll(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HGetAll(key)
	}
	return reply, err
}

// HSet Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
// If field already exists in the hash, it is overwritten.
func (ca *redisCache) HSet(key, field, value string) error {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	err := client.HSet(key, field, value)
	return err
}

// HDel Removes the specified fields from the hash stored at key.
func (ca *redisCache) HDel(key string, field ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.HDel(key, field...)
	return reply, err
}

// HExists Returns if field is an existing field in the hash stored at key
func (ca *redisCache) HExists(key string, field string) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HExist(key, field)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HExist(key, field)
	}
	return reply, err
}

// HSetNX Sets field in the hash stored at key to value, only if field does not yet exist
func (ca *redisCache) HSetNX(key string, field string, value string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.HSetNX(key, field, value)
	return reply, err
}

// HIncrBy Increments the number stored at field in the hash stored at key by increment.
func (ca *redisCache) HIncrBy(key string, field string, increment int) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.HIncrBy(key, field, increment)
	return reply, err
}

// HIncrByFloat Increment the specified field of a hash stored at key, and representing a floating point number, by the specified increment
func (ca *redisCache) HIncrByFloat(key string, field string, increment float64) (float64, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.HIncrByFloat(key, field, increment)
	return reply, err
}

// HKeys Returns all field names in the hash stored at key.
func (ca *redisCache) HKeys(key string) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HKeys(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HKeys(key)
	}
	return reply, err
}

// HLen Returns the number of fields contained in the hash stored at key
func (ca *redisCache) HLen(key string) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HLen(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HLen(key)
	}
	return reply, err
}

// HVals Returns all values in the hash stored at key
func (ca *redisCache) HVals(key string) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.HVals(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.HVals(key)
	}
	return reply, err
}

/*---------- List -----------*/
// BLPop BLPOP is a blocking list pop primitive.
// It is the blocking version of LPOP because it blocks the connection when there are no elements to pop from any of the given lists
func (ca *redisCache) BLPop(key ...interface{}) (map[string]string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.BLPop(key...)
	return reply, err
}

// BRPOP is a blocking list pop primitive
// It is the blocking version of RPOP because it blocks the connection when there are no elements to pop from any of the given lists
func (ca *redisCache) BRPop(key ...interface{}) (map[string]string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.BRPop(key)
	return reply, err
}

// BRPOPLPUSH is a operation like RPOPLPUSH but blocking
func (ca *redisCache) BRPopLPush(source string, destination string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.BRPopLPush(source, destination)
	return reply, err
}

// LIndex return element which subscript is index,
// if index is -1, return last one element of list and so on
func (ca *redisCache) LIndex(key string, index int) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.LIndex(key, index)
}

// LInsert Inserts value in the list stored at key either before or after the reference value pivot.
func (ca *redisCache) LInsert(key string, direction string, pivot string, value string) (int, error) {
	if direction != LInsert_Before && direction != LInsert_After {
		return -1, errors.New("direction only accept BEFORE or AFTER")
	}
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	if direction == LInsert_Before {
		return client.LInsertBefore(key, pivot, value)
	}
	return client.LInsertAfter(key, pivot, value)

}

// LLen return length of list
func (ca *redisCache) LLen(key string) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.LLen(key)
}

// LPop remove and return head element of list
func (ca *redisCache) LPop(key string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.LPop(key)
	return reply, err
}

// LPush Insert all the specified values at the head of the list stored at key
func (ca *redisCache) LPush(key string, value ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.LPush(key, value)
	return reply, err
}

// LPushX insert an element at the head of the list
func (ca *redisCache) LPushX(key string, value string) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.LPushX(key, value)
	return reply, err
}

// LRange Returns the specified elements of the list stored at key
func (ca *redisCache) LRange(key string, start int, stop int) ([]string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.LRange(key, start, stop)
}

// LRem Removes the first count occurrences of elements equal to value from the list stored at key.
func (ca *redisCache) LRem(key string, count int, value string) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.LRem(key, count, value)
	return reply, err
}

// LSet Sets the list element at index to value
func (ca *redisCache) LSet(key string, index int, value string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.LSet(key, index, value)
	return reply, err
}

// LTrim Trim an existing list so that it will contain only the specified range of elements specified
func (ca *redisCache) LTrim(key string, start int, stop int) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.LTrim(key, start, stop)
	return reply, err
}

// RPop Removes and returns the last element of the list stored at key
func (ca *redisCache) RPop(key string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.RPop(key)
	return reply, err
}

// RPopLPush Atomically returns and removes the last element (tail) of the list stored at source, and pushes the element at the first element (head) of the list stored at destination
func (ca *redisCache) RPopLPush(source string, destination string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.RPopLPush(source, destination)
	return reply, err
}

// RPush Insert all the specified values at the tail of the list stored at key.
func (ca *redisCache) RPush(key string, value ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.RPush(key, value...)
	return reply, err
}

// RPushX Inserts value at the tail of the list stored at key, only if key already exists and holds a list
func (ca *redisCache) RPushX(key string, value ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.RPushX(key, value...)
	return reply, err
}

/*---------- Set -----------*/
// SAdd Add the specified members to the set stored at key
func (ca *redisCache) SAdd(key string, member ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.SAdd(key, member)
	return reply, err
}

// SCard Returns the set cardinality (number of elements) of the set stored at key
func (ca *redisCache) SCard(key string) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.SCard(key)
	return reply, err
}

// SDiff Returns the members of the set resulting from the difference between the first set and all the successive sets
func (ca *redisCache) SDiff(key ...interface{}) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SDiff(key...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SDiff(key...)
	}
	return reply, err
}

// SDiffStore This command is equal to SDIFF, but instead of returning the resulting set, it is stored in destination
func (ca *redisCache) SDiffStore(destination string, key ...interface{}) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SDiffStore(destination, key...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SDiffStore(destination, key...)
	}
	return reply, err
}

// SInter Returns the members of the set resulting from the intersection of all the given sets.
func (ca *redisCache) SInter(key ...interface{}) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SInter(key...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SInter(key...)
	}
	return reply, err
}

// SInterStore This command is equal to SINTER, but instead of returning the resulting set, it is stored in destination
func (ca *redisCache) SInterStore(destination string, key ...interface{}) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SInterStore(destination, key...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SInterStore(destination, key...)
	}
	return reply, err
}

// SIsMember Returns if member is a member of the set stored at key.
func (ca *redisCache) SIsMember(key string, member string) (bool, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SIsMember(key, member)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SIsMember(key, member)
	}
	return reply, err
}

// SMembers Returns all the members of the set value stored at key.
func (ca *redisCache) SMembers(key string) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SMembers(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SMembers(key)
	}
	return reply, err
}

// SMove Move member from the set at source to the set at destination
func (ca *redisCache) SMove(source string, destination string, member string) (bool, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.SMove(source, destination, member)
}

// SPop Removes and returns one or more random elements from the set value store at key.
func (ca *redisCache) SPop(key string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.SPop(key)
}

// SRandMember When called with just the key argument, return a random element from the set value stored at key
func (ca *redisCache) SRandMember(key string, count int) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SRandMember(key, count)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SRandMember(key, count)
	}
	return reply, err
}

// SRem Remove the specified members from the set stored at key
func (ca *redisCache) SRem(key string, member ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.SRem(key, member...)
}

// SUnion Returns the members of the set resulting from the union of all the given sets
func (ca *redisCache) SUnion(key ...interface{}) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.SUnion(key...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SUnion(key...)
	}
	return reply, err
}

// SUnionStore This command is equal to SUNION, but instead of returning the resulting set, it is stored in destination
func (ca *redisCache) SUnionStore(destination string, key ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.SUnionStore(destination, key...)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.SUnionStore(destination, key...)
	}
	return reply, err
}

//****************** sorted set 集合 ***********************
// ZAdd Adds all the specified members with the specified scores to the sorted set stored at key
func (ca *redisCache) ZAdd(key string, score int64, member interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.ZAdd(key, score, member)
}

// ZCount Returns the number of elements in the sorted set at key with a score between min and max
func (ca *redisCache) ZCount(key string, min, max int64) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZCount(key, min, max)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZCount(key, min, max)
	}
	return reply, err
}

// ZRem Removes the specified members from the sorted set stored at key. Non existing members are ignored.
func (ca *redisCache) ZRem(key string, member ...interface{}) (int, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.ZRem(key, member...)
}

// ZCard Returns the sorted set cardinality (number of elements) of the sorted set stored at key.
func (ca *redisCache) ZCard(key string) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZCard(key)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZCard(key)
	}
	return reply, err
}

// ZRank Returns the rank of member in the sorted set stored at key, with the scores ordered from low to high
func (ca *redisCache) ZRank(key, member string) (int, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZRank(key, member)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZRank(key, member)
	}
	return reply, err
}

// ZRange Returns the specified range of elements in the sorted set stored at key
func (ca *redisCache) ZRange(key string, start, stop int64) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZRange(key, start, stop)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZRange(key, start, stop)
	}
	return reply, err
}

// ZRangeByScore Returns all the elements in the sorted set at key with a score between min and max (including elements with score equal to min or max).
func (ca *redisCache) ZRangeByScore(key string, start, stop string, isWithScores bool) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZRangeByScore(key, start, stop, isWithScores)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZRangeByScore(key, start, stop, isWithScores)
	}
	return reply, err
}

// ZREVRangeByScore Returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min). In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
func (ca *redisCache) ZREVRangeByScore(key string, max, min string, isWithScores bool) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZREVRangeByScore(key, max, min, isWithScores)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZREVRangeByScore(key, max, min, isWithScores)
	}
	return reply, err
}

// ZRange Returns the specified range of elements in the sorted set stored at key
func (ca *redisCache) ZRevRange(key string, start, stop int64) ([]string, error) {
	client := ca.getReadRedisClient()
	reply, err := client.ZRevRange(key, start, stop)
	if ca.checkConnErrorAndNeedRetry(err) {
		client = ca.getBackupRedis()
		return client.ZRevRange(key, start, stop)
	}
	return reply, err
}

//****************** PUB/SUB *********************
// Publish Posts a message to the given channel.
func (ca *redisCache) Publish(channel string, message interface{}) (int64, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.Publish(channel, message)
}

// Subscribe Subscribes the client to the specified channels
func (ca *redisCache) Subscribe(receive chan Message, channels ...interface{}) error {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	conn := client.GetConn()
	psc := redis.PubSubConn{Conn: conn}
	defer func() {
		conn.Close()
		psc.Unsubscribe(channels...)
	}()

	err := psc.Subscribe(channels...)
	if err != nil {
		return err
	}
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("%s: messages: %s\n", v.Channel, v.Data)
			receive <- Message{Channel: v.Channel, Data: v.Data}
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			continue
		case error:
			return err
		}
	}
}

//****************** lua scripts *********************
// EVAL used to evaluate scripts using the Lua interpreter built into Redis starting from version 2.6.0
func (ca *redisCache) EVAL(script string, argsNum int, arg ...interface{}) (interface{}, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.EVAL(script, argsNum, arg...)
}

//****************** 全局操作 ***********************
// Ping ping command, if success return pong
func (ca *redisCache) Ping() (string, error) {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.Ping()
}

// ClearAll will delete all item in redis cache.
// never error
func (ca *redisCache) ClearAll() error {
	client := internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	client.FlushDB()
	return nil
}

// getReadRedisClient get read mode redis client
func (ca *redisCache) getReadRedisClient() *internal.RedisClient {
	if ca.hystrix.IsHystrix() {
		if ca.backupServerUrl != "" {
			return ca.getBackupRedis()
		}
	}
	if ca.readOnlyServerUrl != "" {
		return ca.getReadOnlyRedis()
	}
	return ca.getDefaultRedis()
}

// getRedisClient get default redis client
func (ca *redisCache) getDefaultRedis() *internal.RedisClient {
	return internal.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
}

func (ca *redisCache) getBackupRedis() *internal.RedisClient {
	return internal.GetRedisClient(ca.backupServerUrl, ca.backupMaxIdle, ca.backupMaxActive)
}

func (ca *redisCache) getReadOnlyRedis() *internal.RedisClient {
	return internal.GetRedisClient(ca.readOnlyServerUrl, ca.readOnlyMaxIdle, ca.readOnlyMaxActive)
}

// checkConnErrorAndNeedRetry check err is Conn error and is need to retry
// if current client is hystrix, no need retry, because in getReadRedisClient already use backUp redis
func (ca *redisCache) checkConnErrorAndNeedRetry(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "no such host") >= 0 ||
		strings.Index(err.Error(), "No connection could be made because the target machine actively refused it") >= 0 ||
		strings.Index(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") >= 0 {
		ca.hystrix.GetCounter().Inc(1)
		//if is hystrix, not to retry, because in getReadRedisClient already use backUp redis
		if ca.hystrix.IsHystrix() {
			return false
		}
		if ca.backupServerUrl == "" {
			return false
		}
		return true
	}
	return false
}

// checkRedisAlive check redis is alive use ping
// if set readonly redis, check readonly redis
// if not set readonly redis, check default redis
func (ca *redisCache) checkRedisAlive() bool {
	isAlive := false
	var redisClient *internal.RedisClient
	if ca.readOnlyServerUrl != "" {
		redisClient = ca.getReadOnlyRedis()
	} else {
		redisClient = ca.getDefaultRedis()
	}
	for i := 0; i <= 5; i++ {
		reply, err := redisClient.Ping()
		//fmt.Println(time.Now(), "checkAliveDefaultRedis Ping", reply, err)
		if err != nil {
			isAlive = false
			break
		}
		if reply != "PONG" {
			isAlive = false
			break
		}
		isAlive = true
		continue
	}
	return isAlive
}
