package redis

import (
	"github.com/devfeel/cache/internal"
	"strconv"
)

var (
	ZeroInt64 int64 = 0
)

// RedisCache is redis cache adapter.
// it contains serverIp for redis conn.
type redisCache struct {
	serverIp string //connection string, like "10.0.1.11:6379"
}

// NewRedisCache returns a new *RedisCache.
func NewRedisCache(serverIp string) *redisCache {
	cache := redisCache{serverIp: serverIp}
	return &cache
}

// Exists check item exist in redis cache.
func (ca *redisCache) Exists(key string) (bool, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	exists, err := redisClient.Exists(key)
	return exists, err
}

// Incr increase int64 counter in redis cache.
func (ca *redisCache) Incr(key string) (int64, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	val, err := redisClient.INCR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Decr decrease counter in redis cache.
func (ca *redisCache) Decr(key string) (int64, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	val, err := redisClient.DECR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Get cache from redis cache.
// if non-existed or expired, return nil.
func (ca *redisCache) Get(key string) (interface{}, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	reply, err := redisClient.GetObj(key)
	return reply, err
}

//  returns value string format by given key
// if non-existed or expired, return "".
func (ca *redisCache) GetString(key string) (string, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	reply, err := redisClient.Get(key)
	return reply, err
}

//  returns value int format by given key
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

//  returns value int64 format by given key
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
	redisClient := internal.GetRedisClient(ca.serverIp)
	_, err := redisClient.SetWithExpire(key, value, ttl)
	return err
}

// Delete item in redis cacha.
// if not exists, we think it's success
func (ca *redisCache) Delete(key string) error {
	redisClient := internal.GetRedisClient(ca.serverIp)
	_, err := redisClient.Del(key)
	return err
}

// ClearAll will delete all item in redis cache.
// never error
func (ca *redisCache) ClearAll() error {
	redisClient := internal.GetRedisClient(ca.serverIp)
	redisClient.FlushDB()
	return nil
}

//Returns the value associated with field in the hash stored at key.
func (ca *redisCache) HGet(hashID, field string) (string, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	return redisClient.HGet(hashID, field)
}

//Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
//If field already exists in the hash, it is overwritten.
func (ca *redisCache) HSet(hashID, field, val string) error {
	redisClient := internal.GetRedisClient(ca.serverIp)
	return redisClient.HSet(hashID, field, val)
}

//Insert all the specified values at the head of the list stored at key
func (ca *redisCache) LPush(key string, val string) (int64, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	return redisClient.LPush(key, val)
}

//BRPOP is a blocking list pop primitive
func (ca *redisCache) BRPop(key string) (string, error) {
	redisClient := internal.GetRedisClient(ca.serverIp)
	return redisClient.BRPop(key)
}
