package redis

import (
	"github.com/devfeel/cache/internal"		//internal目录 不允许其他包调用, commit时候改回来
	"strconv"
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
	serverUrl string //connection string, like "redis://:password@10.0.1.11:6379/0"
}

// NewRedisCache returns a new *RedisCache.
func NewRedisCache(serverUrl string) *redisCache {
	cache := redisCache{serverUrl: serverUrl}
	return &cache
}

// Exists check item exist in redis cache.
func (ca *redisCache) Exists(key string) (bool, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	exists, err := client.Exists(key)
	return exists, err
}

// Incr increase int64 counter in redis cache.
func (ca *redisCache) Incr(key string) (int64, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	val, err := client.INCR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Decr decrease counter in redis cache.
func (ca *redisCache) Decr(key string) (int64, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	val, err := client.DECR(key)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Get cache from redis cache.
// if non-existed or expired, return nil.
func (ca *redisCache) Get(key string) (interface{}, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	reply, err := client.GetObj(key)
	return reply, err
}

// GetString returns value string format by given key
// if non-existed or expired, return "".
func (ca *redisCache) GetString(key string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	reply, err := client.Get(key)
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
	client := internal.GetRedisClient(ca.serverUrl)
	_, err := client.SetWithExpire(key, value, ttl)
	return err
}

// Delete item in redis cacha.
// if not exists, we think it's success
func (ca *redisCache) Delete(key string) error {
	client := internal.GetRedisClient(ca.serverUrl)
	_, err := client.Del(key)
	return err
}

// ClearAll will delete all item in redis cache.
// never error
func (ca *redisCache) ClearAll() error {
	client := internal.GetRedisClient(ca.serverUrl)
	client.FlushDB()
	return nil
}

// GetJsonObj get obj with SetJsonObj key
func (ca *redisCache) GetJsonObj(key string, result interface{})error {
	client := internal.GetRedisClient(ca.serverUrl)
	err := client.GetJsonObj(key, result)
	return err
}

// SetJsonObj set obj use json encode string
func (ca *redisCache) SetJsonObj(key string, val interface{}) (interface{}, error){
	client := internal.GetRedisClient(ca.serverUrl)
	reply, err := client.SetJsonObj(key, val)
	return reply, err
}

/*---------- Hash -----------*/
// HGet Returns the value associated with field in the hash stored at key.
func (ca *redisCache) HGet(key, field string) (string, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HGet(key, field)
}
// HGetAll Returns all fields and values of the hash stored at key
func (ca *redisCache) HGetAll(key string) (map[string]string, error) {
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HGetAll(key)
}
// HSet Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
// If field already exists in the hash, it is overwritten.
func (ca *redisCache) HSet(key, field, value string) error {
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HSet(key, field, value)
}
// HDel Removes the specified fields from the hash stored at key.
func (ca *redisCache) HDel(key string, field ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HDel(key, field...)
}
// HExists Returns if field is an existing field in the hash stored at key
func (ca *redisCache) HExists (key string, field string) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HExist(key, field)
}
// HSetNX Sets field in the hash stored at key to value, only if field does not yet exist
func (ca *redisCache) HSetNX(key string, field string, value string) (string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HSetNX(key, field, value)
}
// HIncrBy Increments the number stored at field in the hash stored at key by increment.
func (ca *redisCache) HIncrBy(key string, field string, increment int) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HIncrBy(key, field, increment)
}
// HIncrByFloat Increment the specified field of a hash stored at key, and representing a floating point number, by the specified increment
func (ca *redisCache) HIncrByFloat(key string, field string, increment float64) (float64, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HIncrByFloat(key, field, increment)
}
// HKeys Returns all field names in the hash stored at key.
func (ca *redisCache) HKeys(key string) ([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HKeys(key)
}
// HLen Returns the number of fields contained in the hash stored at key
func (ca *redisCache) HLen(key string) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HLen(key)
}
// HVals Returns all values in the hash stored at key
func (ca *redisCache) HVals(key string) ([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.HVals(key)
}

/*---------- List -----------*/
// BLPop BLPOP is a blocking list pop primitive.
// It is the blocking version of LPOP because it blocks the connection when there are no elements to pop from any of the given lists
func (ca *redisCache) BLPop(key ...interface{})(map[string]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.BLPop(key...)
}
// BRPOP is a blocking list pop primitive
// It is the blocking version of RPOP because it blocks the connection when there are no elements to pop from any of the given lists
func (ca *redisCache) BRPop(key ...interface{}) (map[string]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.BRPop(key)
}
// BRPOPLPUSH is a operation like RPOPLPUSH but blocking
func (ca *redisCache) BRPopLPush(source string, destination string)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.BRPopLPush(source, destination)
}
// LIndex return element which subscript is index,
// if index is -1, return last one element of list and so on
func (ca *redisCache) LIndex(key string, index int)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LIndex(key, index)
}
// LInsert Inserts value in the list stored at key either before or after the reference value pivot.
func (ca *redisCache) LInsert(key string, direction string, pivot string, value string)(int, error){
	if direction != LInsert_Before && direction != LInsert_After {
		return -1, errors.New("direction only accept BEFORE or AFTER")
	}
	client := internal.GetRedisClient(ca.serverUrl)
	if direction == LInsert_Before {
		return client.LInsertBefore(key, pivot, value)
	}
	return client.LInsertAfter(key, pivot, value)

}
// LLen return length of list
func (ca *redisCache) LLen(key string) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LLen(key)
}
// LPop remove and return head element of list
func (ca *redisCache) LPop(key string)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LPop(key)
}
// LPush Insert all the specified values at the head of the list stored at key
func (ca *redisCache) LPush(key string, value ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LPush(key, value)
}
// LPushX insert an element at the head of the list
func (ca *redisCache) LPushX(key string, value string)(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LPushX(key, value)
}
// LRange Returns the specified elements of the list stored at key
func (ca *redisCache) LRange(key string, start int, stop int)([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LRange(key, start, stop)
}
// LRem Removes the first count occurrences of elements equal to value from the list stored at key.
func (ca *redisCache) LRem(key string, count int, value string)(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LRem(key, count, value)
}
// LSet Sets the list element at index to value
func (ca *redisCache) LSet(key string, index int, value string)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LSet(key, index, value)
}
// LTrim Trim an existing list so that it will contain only the specified range of elements specified
func (ca *redisCache) LTrim(key string, start int, stop int)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.LTrim(key, start, stop)
}
// RPop Removes and returns the last element of the list stored at key
func (ca *redisCache) RPop(key string)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.RPop(key)
}
// RPopLPush Atomically returns and removes the last element (tail) of the list stored at source, and pushes the element at the first element (head) of the list stored at destination
func (ca *redisCache) RPopLPush(source string, destination string)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.RPopLPush(source, destination)
}
// RPush Insert all the specified values at the tail of the list stored at key.
func (ca *redisCache) RPush(key string, value ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.RPush(key, value...)
}
// RPushX Inserts value at the tail of the list stored at key, only if key already exists and holds a list
func (ca *redisCache) RPushX(key string, value ...interface{}) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.RPushX(key, value...)
}

/*---------- Set -----------*/
// SAdd Add the specified members to the set stored at key
func (ca *redisCache) SAdd(key string, member ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SAdd(key, member)
}
// SCard Returns the set cardinality (number of elements) of the set stored at key
func (ca *redisCache) SCard(key string) (int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SCard(key)
}
// SDiff Returns the members of the set resulting from the difference between the first set and all the successive sets
func (ca *redisCache) SDiff(key ...interface{})([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SDiff(key...)
}
// SDiffStore This command is equal to SDIFF, but instead of returning the resulting set, it is stored in destination
func (ca *redisCache) SDiffStore(destination string, key ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SDiffStore(destination, key...)
}
// SInter Returns the members of the set resulting from the intersection of all the given sets.
func (ca *redisCache) SInter(key ...interface{})([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SInter(key...)
}
// SInterStore This command is equal to SINTER, but instead of returning the resulting set, it is stored in destination
func (ca *redisCache) SInterStore(destination string, key ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SInterStore(destination, key...)
}
// SIsMember Returns if member is a member of the set stored at key.
func (ca *redisCache) SIsMember(key string, member string)(bool, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SIsMember(key, member)
}
// SMembers Returns all the members of the set value stored at key.
func (ca *redisCache) SMembers(key string)([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SMembers(key)
}
// SMove Move member from the set at source to the set at destination
func (ca *redisCache) SMove(source string, destination string, member string)(bool, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SMove(source, destination, member)
}
// SPop Removes and returns one or more random elements from the set value store at key.
func (ca *redisCache) SPop(key string)(string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SPop(key)
}
// SRandMember When called with just the key argument, return a random element from the set value stored at key
func (ca *redisCache) SRandMember(key string, count int)([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SRandMember(key, count)
}
// SRem Remove the specified members from the set stored at key
func (ca *redisCache) SRem(key string, member ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SRem(key, member...)
}
// SUnion Returns the members of the set resulting from the union of all the given sets
func (ca *redisCache)  SUnion(key ...interface{})([]string, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SUnion(key...)
}
// SUnionStore This command is equal to SUNION, but instead of returning the resulting set, it is stored in destination
func (ca *redisCache)  SUnionStore(destination string, key ...interface{})(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.SUnionStore(destination, key...)
}

//****************** sorted set 集合 ***********************
// ZAdd Adds all the specified members with the specified scores to the sorted set stored at key
func (ca *redisCache) ZAdd(key string, score int64, member interface{})(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.ZAdd(key, score, member)
}

// ZCount Returns the number of elements in the sorted set at key with a score between min and max
func (ca *redisCache) ZCount(key string, min, max int64)(int, error){
	client := internal.GetRedisClient(ca.serverUrl)
	return client.ZCount(key, min, max)
}