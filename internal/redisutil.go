// redisclient
package internal

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"sync"
)

type RedisClient struct {
	pool    *redis.Pool
	Address string
}

var OnConnError func()

var (
	redisMap map[string]*RedisClient
	mapMutex *sync.RWMutex
)

const (
	defaultTimeout = 60 * 10 //默认10分钟
)

func init() {
	redisMap = make(map[string]*RedisClient)
	mapMutex = new(sync.RWMutex)
}

// 重写生成连接池方法
// redisURL: connection string, like "redis://:password@10.0.1.11:6379/0"
func newPool(redisURL string, maxIdle, maxActive int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   maxIdle,
		MaxActive: maxActive, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			return c, err
		},
	}
}

//获取指定Address及连接池设置的RedisClient
func GetRedisClient(address string, maxIdle int, maxActive int) *RedisClient {
	var redis *RedisClient
	var mok bool
	mapMutex.RLock()
	redis, mok = redisMap[address]
	mapMutex.RUnlock()
	if !mok {
		redis = &RedisClient{Address: address, pool: newPool(address, maxIdle, maxActive)}
		mapMutex.Lock()
		redisMap[address] = redis
		mapMutex.Unlock()
	}
	return redis
}

//获取指定key的内容, interface{}
func (rc *RedisClient) GetObj(key string) (interface{}, error) {
	// 从连接池里面获得一个连接
	conn := rc.pool.Get()
	// 连接完关闭，其实没有关闭，是放回池里，也就是队列里面，等待下一个重用
	defer conn.Close()
	reply, errDo := innerDo(conn, "GET", key)
	return reply, errDo
}

//获取指定key的内容, string
func (rc *RedisClient) Get(key string) (string, error) {
	val, err := redis.String(rc.GetObj(key))
	return val, err
}

//检查指定key是否存在
func (rc *RedisClient) Exists(key string) (bool, error) {
	// 从连接池里面获得一个连接
	conn := rc.pool.Get()
	// 连接完关闭，其实没有关闭，是放回池里，也就是队列里面，等待下一个重用
	defer conn.Close()
	reply, errDo := innerDo(conn, "EXISTS", key)
	if errDo == nil && reply == nil {
		return false, nil
	}
	val, err := redis.Int(reply, errDo)
	return val > 0, err
}

//删除指定key
func (rc *RedisClient) Del(key ...interface{}) (int, error) {
	// 从连接池里面获得一个连接
	conn := rc.pool.Get()
	// 连接完关闭，其实没有关闭，是放回池里，也就是队列里面，等待下一个重用
	defer conn.Close()
	reply, errDo := conn.Do("DEL", key...)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

//对存储在指定key的数值执行原子的加1操作
func (rc *RedisClient) INCR(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := innerDo(conn, "INCR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

//对存储在指定key的数值执行原子的减1操作
func (rc *RedisClient) DECR(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := innerDo(conn, "DECR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (rc *RedisClient) Set(key string, val interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := innerDo(conn, "SET", key, val)
	return val, err
}

// SetWithExpire 设置指定key的内容
func (rc *RedisClient) SetWithExpire(key string, val interface{}, timeOutSeconds int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := innerDo(conn, "SET", key, val, "EX", timeOutSeconds)
	return val, err
}

// SetNX  将 key 的值设为 value ，当且仅当 key 不存在。
// 若给定的 key 已经存在，则 SETNX 不做任何动作。 成功返回1, 失败返回0
func (rc *RedisClient) SetNX(key, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int(innerDo(conn, "SETNX", key, value))
	return val, err
}

// Expire 设置指定key的过期时间
func (rc *RedisClient) Expire(key string, timeOutSeconds int) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "EXPIRE", key, timeOutSeconds))
	return val, err
}

// GetJsonObj get obj with SetJsonObj key
func (rc *RedisClient) GetJsonObj(key string, result interface{}) error {
	jsonStr, err := redis.String(rc.GetObj(key))
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(jsonStr), result)
	return err
}

// SetJsonObj set obj use json encode string
func (rc *RedisClient) SetJsonObj(key string, val interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	jsonStr, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	reply, err := redis.String(innerDo(conn, "SET", key, jsonStr))
	return reply, err
}

//删除当前数据库里面的所有数据
//这个命令永远不会出现失败
func (rc *RedisClient) FlushDB() {
	conn := rc.pool.Get()
	defer conn.Close()
	conn.Do("FLUSHALL")
}

//****************** hash 哈希表 ***********************

//获取指定hashset的所有内容
func (rc *RedisClient) HGetAll(hashID string) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, err := redis.StringMap(innerDo(conn, "HGETALL", hashID))
	return reply, err
}

//获取指定hashset的内容
func (rc *RedisClient) HGet(hashID string, field string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("HGET", hashID, field)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

// HMGet 返回 key 指定的哈希集中指定字段的值
func (rc *RedisClient) HMGet(hashID string, field ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{hashID}, field...)
	reply, err := redis.Strings(innerDo(conn, "HMGET", args...))
	return reply, err
}

//设置指定hashset的内容
func (rc *RedisClient) HSet(hashID string, field string, val string) error {
	conn := rc.pool.Get()
	defer conn.Close()
	_, err := innerDo(conn, "HSET", hashID, field, val)
	return err
}

func (rc *RedisClient) HSetNX(hashID string, field string, val string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, err := redis.String(innerDo(conn, "HSETNX", hashID, field, val))
	return reply, err
}

func (rc *RedisClient) HDel(key string, field ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, field...)
	val, err := redis.Int(innerDo(conn, "HDEL", args...))
	return val, err
}

func (rc *RedisClient) HExist(key string, field string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "HEXISTS", key, field))
	return val, err
}
func (rc *RedisClient) HIncrBy(key string, field string, increment int) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "HINCRBY", key, field, increment))
	return val, err
}

func (rc *RedisClient) HIncrByFloat(key string, field string, increment float64) (float64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Float64(innerDo(conn, "HINCRBYFLOAT", key, field, increment))
	return val, err
}

func (rc *RedisClient) HKeys(key string) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "HKEYS", key))
	return val, err
}

// HLen 返回哈希表 key 中域的数量, 当 key 不存在时，返回0
func (rc *RedisClient) HLen(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "HLEN", key))
	return val, err
}

// HVals 返回哈希表 key 中所有域的值, 当 key 不存在时，返回空
func (rc *RedisClient) HVals(key string) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "HVALS", key))
	return val, err
}

//****************** list 链表 ***********************

//将所有指定的值插入到存于 key 的列表的头部
func (rc *RedisClient) LPush(key string, value ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	ret, err := redis.Int(innerDo(conn, "LPUSH", key, value))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}

func (rc *RedisClient) LPushX(key string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Int(innerDo(conn, "LPUSHX", key, value))
	return resp, err
}

func (rc *RedisClient) LRange(key string, start int, stop int) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Strings(innerDo(conn, "LRANGE", key, start, stop))
	return resp, err
}

func (rc *RedisClient) LRem(key string, count int, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Int(innerDo(conn, "LREM", key, count, value))
	return resp, err
}

func (rc *RedisClient) LSet(key string, index int, value string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(innerDo(conn, "LSET", key, index, value))
	return resp, err
}

func (rc *RedisClient) LTrim(key string, start int, stop int) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(innerDo(conn, "LTRIM", key, start, stop))
	return resp, err
}

func (rc *RedisClient) RPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(innerDo(conn, "RPOP", key))
	return resp, err
}

func (rc *RedisClient) RPush(key string, value ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, value...)
	resp, err := redis.Int(innerDo(conn, "RPUSH", args...))
	return resp, err
}

func (rc *RedisClient) RPushX(key string, value ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, value...)
	resp, err := redis.Int(innerDo(conn, "RPUSHX", args...))
	return resp, err
}

func (rc *RedisClient) RPopLPush(source string, destination string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(innerDo(conn, "RPOPLPUSH", source, destination))
	return resp, err
}

func (rc *RedisClient) BLPop(key ...interface{}) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(innerDo(conn, "BLPOP", key, defaultTimeout))
	return val, err
}

//删除，并获得该列表中的最后一个元素，或阻塞，直到有一个可用
func (rc *RedisClient) BRPop(key ...interface{}) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(innerDo(conn, "BRPOP", key, defaultTimeout))
	return val, err
}

func (rc *RedisClient) BRPopLPush(source string, destination string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(innerDo(conn, "BRPOPLPUSH", source, destination))
	return val, err
}

func (rc *RedisClient) LIndex(key string, index int) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(innerDo(conn, "LINDEX", key, index))
	return val, err
}

func (rc *RedisClient) LInsertBefore(key string, pivot string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "LINSERT", key, "BEFORE", pivot, value))
	return val, err
}

func (rc *RedisClient) LInsertAfter(key string, pivot string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "LINSERT", key, "AFTER", pivot, value))
	return val, err
}

func (rc *RedisClient) LLen(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "LLEN", key))
	return val, err
}

func (rc *RedisClient) LPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(innerDo(conn, "LPOP", key))
	return val, err
}

//****************** set 集合 ***********************

// SAdd 将一个或多个 member 元素加入到集合 key 当中，已经存在于集合的 member 元素将被忽略。
// 假如 key 不存在，则创建一个只包含 member 元素作成员的集合。
func (rc *RedisClient) SAdd(key string, member ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, member...)
	val, err := redis.Int(innerDo(conn, "SADD", args...))
	return val, err
}

// SCard 返回集合 key 的基数(集合中元素的数量)。
// 返回值：
// 集合的基数。
// 当 key 不存在时，返回 0
func (rc *RedisClient) SCard(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "SCARD", key))
	return val, err
}

// SPop 移除并返回集合中的一个随机元素。
// 如果只想获取一个随机元素，但不想该元素从集合中被移除的话，可以使用 SRANDMEMBER 命令。
// count 为 返回的随机元素的数量
func (rc *RedisClient) SPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(innerDo(conn, "SPOP", key))
	return val, err
}

// SRandMember 如果命令执行时，只提供了 key 参数，那么返回集合中的一个随机元素。
// 该操作和 SPOP 相似，但 SPOP 将随机元素从集合中移除并返回，而 SRANDMEMBER 则仅仅返回随机元素，而不对集合进行任何改动。
// count 为 返回的随机元素的数量
func (rc *RedisClient) SRandMember(key string, count int) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "SRANDMEMBER", key, count))
	return val, err
}

// SRem 移除集合 key 中的一个或多个 member 元素，不存在的 member 元素会被忽略。
// 当 key 不是集合类型，返回一个错误。
// 在 Redis 2.4 版本以前， SREM 只接受单个 member 值。
func (rc *RedisClient) SRem(key string, member ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, member...)
	val, err := redis.Int(innerDo(conn, "SREM", args...))
	return val, err
}

func (rc *RedisClient) SDiff(key ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "SDIFF", key...))
	return val, err
}

func (rc *RedisClient) SDiffStore(destination string, key ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{destination}, key...)
	val, err := redis.Int(innerDo(conn, "SDIFFSTORE", args...))
	return val, err
}

func (rc *RedisClient) SInter(key ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "SINTER", key...))
	return val, err
}

func (rc *RedisClient) SInterStore(destination string, key ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{destination}, key...)
	val, err := redis.Int(innerDo(conn, "SINTERSTORE", args...))
	return val, err
}

func (rc *RedisClient) SIsMember(key string, member string) (bool, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Bool(innerDo(conn, "SISMEMBER", key, member))
	return val, err
}

func (rc *RedisClient) SMembers(key string) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "SMEMBERS", key))
	return val, err
}

// smove is a atomic operate
func (rc *RedisClient) SMove(source string, destination string, member string) (bool, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Bool(innerDo(conn, "SMOVE", source, destination, member))
	return val, err
}

func (rc *RedisClient) SUnion(key ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(innerDo(conn, "SUNION", key...))
	return val, err
}

func (rc *RedisClient) SUnionStore(destination string, key ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{destination}, key...)
	val, err := redis.Int(innerDo(conn, "SUNIONSTORE", args))
	return val, err
}

//****************** sorted set 集合 ***********************

// ZAdd 将所有指定成员添加到键为key有序集合（sorted set）里面。 添加时可以指定多个分数/成员（score/member）对。
// 如果指定添加的成员已经是有序集合里面的成员，则会更新改成员的分数（scrore）并更新到正确的排序位置
func (rc *RedisClient) ZAdd(key string, score int64, member interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, score, member)
	val, err := redis.Int(innerDo(conn, "ZADD", args...))
	return val, err
}

// ZCount 返回有序集key中，score值在min和max之间(默认包括score值等于min或max)的成员
func (rc *RedisClient) ZCount(key string, min, max int64) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, min, max)
	val, err := redis.Int(innerDo(conn, "ZCOUNT", args...))
	return val, err
}

// ZRem 从排序的集合中删除一个或多个成员
// 当key存在，但是其不是有序集合类型，就返回一个错误。
func (rc *RedisClient) ZRem(key string, member ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, member...)
	val, err := redis.Int(innerDo(conn, "ZREM", args...))
	return val, err
}

// ZCard 返回key的有序集元素个数
func (rc *RedisClient) ZCard(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key})
	val, err := redis.Int(innerDo(conn, "ZCARD", args...))
	return val, err
}

// ZRank 返回有序集key中成员member的排名
func (rc *RedisClient) ZRank(key, member string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, member)
	val, err := redis.Int(innerDo(conn, "ZRANK", args...))
	return val, err
}

// ZRange Returns the specified range of elements in the sorted set stored at key
func (rc *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, start, stop)
	val, err := redis.Strings(innerDo(conn, "ZRANGE", args...))
	return val, err
}

// ZRangeByScore Returns all the elements in the sorted set at key with a score between min and max (including elements with score equal to min or max).
func (rc *RedisClient) ZRangeByScore(key string, start, stop string, isWithScores bool) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, start, stop)
	if isWithScores {
		args = append(args, "WITHSCORES")
	}
	val, err := redis.Strings(innerDo(conn, "ZRANGEBYSCORE", args...))
	return val, err
}

// ZREVRangeByScore Returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min). In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
func (rc *RedisClient) ZREVRangeByScore(key string, max, min string, isWithScores bool) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, max, min)
	if isWithScores {
		args = append(args, "WITHSCORES")
	}
	val, err := redis.Strings(innerDo(conn, "ZREVRANGEBYSCORE", args...))
	return val, err
}

// ZRange Returns the specified range of elements in the sorted set stored at key
func (rc *RedisClient) ZRevRange(key string, start, stop int64) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, start, stop)
	val, err := redis.Strings(innerDo(conn, "ZREVRANGE", args...))
	return val, err
}

//****************** PUB/SUB *********************

// Publish 将信息 message 发送到指定的频道 channel
func (rc *RedisClient) Publish(channel string, message interface{}) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	var args []interface{}
	args = append([]interface{}{channel}, message)
	val, err := redis.Int64(innerDo(conn, "PUBLISH", args...))
	return val, err
}

//****************** lua scripts *********************
// EVAL 使用内置的 Lua 解释器
func (rc *RedisClient) EVAL(script string, argsNum int, arg ...interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	var args []interface{}
	if len(arg) > 0 {
		args = append([]interface{}{script, argsNum}, arg...)
	} else {
		args = append([]interface{}{script, argsNum})
	}
	args = append([]interface{}{script, argsNum}, arg...)
	val, err := innerDo(conn, "EVAL", args...)
	return val, err
}

//****************** 全局操作 ***********************
// DBSize 返回当前数据库的 key 的数量
func (rc *RedisClient) DBSize() (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(innerDo(conn, "DBSIZE"))
	return val, err
}

// Ping ping command, if success return pong
func (rc *RedisClient) Ping() (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	var args []interface{}
	val, err := redis.String(innerDo(conn, "PING", args...))
	return val, err
}

// GetConn 返回一个从连接池获取的redis连接,
// 需要手动释放redis连接
func (rc *RedisClient) GetConn() redis.Conn {
	return rc.pool.Get()
}

// Do sends a command to the server and returns the received reply.
func innerDo(conn redis.Conn, commandName string, args ...interface{}) (interface{}, error) {
	reply, err := conn.Do(commandName, args...)
	return reply, err
}
