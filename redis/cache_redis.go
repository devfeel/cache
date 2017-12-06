package redis

import (
	//"github.com/devfeel/cache/internal"		//internal目录 不允许其他包调用, commit时候改回来
	"strconv"

	"github.com/chacha923/cache/internal"
	"errors"
)

var (
	ZeroInt64 int64 = 0
)
const (
	LInsert_Before = "BEFORE"
	LInsert_After = "AFTER"
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
	client := internal.GetRedisClient(ca.serverIp)
	exists, err := client.Exists(key)
	return exists, err
}

// Incr increase int64 counter in redis cache.
func (ca *redisCache) Incr(key string) (int64, error) {
	client := internal.GetRedisClient(ca.serverIp)
	val, err := client.INCR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Decr decrease counter in redis cache.
func (ca *redisCache) Decr(key string) (int64, error) {
	client := internal.GetRedisClient(ca.serverIp)
	val, err := client.DECR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Get cache from redis cache.
// if non-existed or expired, return nil.
func (ca *redisCache) Get(key string) (interface{}, error) {
	client := internal.GetRedisClient(ca.serverIp)
	reply, err := client.GetObj(key)
	return reply, err
}

//  returns value string format by given key
// if non-existed or expired, return "".
func (ca *redisCache) GetString(key string) (string, error) {
	client := internal.GetRedisClient(ca.serverIp)
	reply, err := client.Get(key)
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
func (ca *redisCache) Set(key string, value interface{}, ttl int) error {
	client := internal.GetRedisClient(ca.serverIp)
	_, err := client.SetWithExpire(key, value, ttl)
	return err
}

// Delete item in redis cacha.
// if not exists, we think it's success
func (ca *redisCache) Delete(key string) error {
	client := internal.GetRedisClient(ca.serverIp)
	_, err := client.Del(key)
	return err
}

// ClearAll will delete all item in redis cache.
// never error
func (ca *redisCache) ClearAll() error {
	client := internal.GetRedisClient(ca.serverIp)
	client.FlushDB()
	return nil
}

//Returns the value associated with field in the hash stored at key.
func (ca *redisCache) HGet(key, field string) (string, error) {
	client := internal.GetRedisClient(ca.serverIp)
	return client.HGet(key, field)
}

//Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
//If field already exists in the hash, it is overwritten.
func (ca *redisCache) HSet(key, field, value string) error {
	client := internal.GetRedisClient(ca.serverIp)
	return client.HSet(key, field, value)
}

func (ca *redisCache) HDel(key string, field ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HDel(key, field...)
}

func (ca *redisCache) HExists (key string, field string) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HExist(key, field)
}

func (ca *redisCache) HSetNX(key string, field string, value string) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HSetNX(key, field, value)
}
func (ca *redisCache) HIncrBy(key string, field string, increment int) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HIncrBy(key, field, increment)
}

func (ca *redisCache) HIncrByFloat(key string, field string, increment float64) (float64, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HIncrByFloat(key, field, increment)
}

func (ca *redisCache) HKeys(key string) ([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HKeys(key)
}

func (ca *redisCache) HLen(key string) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HLen(key)
}
func (ca *redisCache) HVals(key string) ([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.HVals(key)
}

func (ca *redisCache) BLPop(key ...interface{})(map[string]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.BLPop(key)
}

//BRPOP is a blocking list pop primitive
func (ca *redisCache) BRPop(key ...interface{}) (map[string]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.BRPop(key)
}

//BRPOPLPUSH is a operation like RPOPLPUSH but blocking
func (ca *redisCache) BRPopLPush(source string, destination string)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.BRPopLPush(source, destination)
}

//return element which subscript is index,
// if index is -1, return last one element of list and so on
func (ca *redisCache) LIndex(key string, index int)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LIndex(key, index)
}

//LINSERT key BEFORE|AFTER pivot value
//insert a new element after one element of list
func (ca *redisCache) LInsert(key string, direction string, pivot string, value string)(int, error){
	if direction != LInsert_Before && direction != LInsert_After {
		return -1, errors.New("direction only accept BEFORE or AFTER")
	}
	client := internal.GetRedisClient(ca.serverIp)
	if direction == LInsert_Before {
		return client.LInsertBefore(key, pivot, value)
	}
	return client.LInsertAfter(key, pivot, value)

}



//return length of list
func (ca *redisCache) LLen(key string) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LLen(key)
}
//remove and return head element of list
func (ca *redisCache) LPop(key string)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LPop(key)
}
//Insert all the specified values at the head of the list stored at key
func (ca *redisCache) LPush(key string, value ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LPush(key, value)
}
//insert an element at the head of the list
func (ca *redisCache) LPushX(key string, value string)(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LPushX(key, value)
}
//LRANGE key start stop
func (ca *redisCache) LRange(key string, start int, stop int)([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LRange(key, start, stop)
}

func (ca *redisCache) LRem(key string, count int, value string)(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LRem(key, count, value)
}

func (ca *redisCache) LSet(key string, index int, value string)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LSet(key, index, value)
}

func (ca *redisCache) LTrim(key string, start int, stop int)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.LTrim(key, start, stop)
}

func (ca *redisCache) RPop(key string)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.RPop(key)
}

func (ca *redisCache) RPopLPush(source string, destination string)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.RPopLPush(source, destination)
}

//RPUSH key value [value ...]
func (ca *redisCache) RPush(key string, value ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.RPush(key, value...)
}

//push a value to list only if list is exist and return length of list after push
// or return 0
func (ca *redisCache) RPushX(key string, value string) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.RPushX(key, value)
}


func (ca *redisCache) SAdd(key string, member ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SAdd(key, member)
}
func (ca *redisCache) SCard(key string) (int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SCard(key)
}
func (ca *redisCache) SDiff(key ...interface{})([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SDiff(key...)
}

func (ca *redisCache) SDiffStore(destination string, key ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SDiffStore(destination, key...)
}
func (ca *redisCache) SInter(key ...interface{})([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SInter(key...)
}
func (ca *redisCache) SInterStore(destination string, key ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SInterStore(destination, key...)
}
func (ca *redisCache) SIsMember(key string, member string)(bool, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SIsMember(key, member)
}
func (ca *redisCache) SMembers(key string)([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SMembers(key)
}
func (ca *redisCache) SMove(source string, destination string, member string)(bool, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SMove(source, destination, member)
}
func (ca *redisCache) SPop(key string, count int)(string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SPop(key, count)
}
func (ca *redisCache) SRandMember(key string, count int)([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SRandMember(key, count)
}
func (ca *redisCache) SRem(key string, member ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SRem(key, member...)
}
func (ca *redisCache)  SUnion(key ...string)([]string, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SUnion(key)
}
func (ca *redisCache)  SUnionStore(destination string, key ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverIp)
	return client.SUnionStore(destination, key...)
}
