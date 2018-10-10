## cache版本记录：

#### Version 0.7.2
* New Feature: Add Command - RedisCache.ZREVRangeByScore(key string, max, min string, isWithScores bool)([]string, error)
-  2018-10-10 15:00

#### Version 0.7.1
- New Feature: Add hystrix module
- New Feature: RedisCache ReadOnly\Backup add hystrix check
- Detail:
-   1、hystrix only use to read command
-   2、if set ReadOnlyServer, hystrix only check readonly server conn
-   3、if only set DefaultServer and BackupServer, hystrix only check default server conn
-  2018-08-17 15:00

#### Version 0.7
- New Feature: RedisCache add BackupServer\ReadOnlyServer
- you can use SetReadOnlyServer or SetBackupServer to set redis info
- Detail:
-   1、if set ReadOnlyServer, all read command will use this server config
-   2、BackupServer only can use to read command
-   3、if set BackupServer, if read command conn server failed, will auto use this config
- Example:
    ``` golang
    redisServer := "redis://192.168.8.175:6329/0"
    readOnlyServer := "redis://192.168.8.175:6339/0"
    backupRedisServer := "redis://192.168.8.175:6379/0"
    redisCache := cache.GetRedisCachePoolConf(redisServer, 10, 100)
    redisCache.SetReadOnlyServer(readOnlyServer, 10, 100)
    redisCache.SetBackupServer(backupRedisServer, 10, 100)
    ```
-  2018-08-16 11:00

#### Version 0.6.6
* New Command: RedisCache.Publish(channel string, message interface{})(int64, error)
* New Command: RedisCache.Subscribe(receive chan redis.Message, channels ...interface{})error
* Example:
    ``` golang
    func TestRedisCache_Subscribe(t *testing.T) {
        receive := make(chan Message, 1000)
        err := rc.Subscribe(receive, "channel-test")
        if err != nil{
            t.Error("TestRedisCache_Subscribe error", err)
        }
        for{
            select{
                case msg := <- receive:
                    fmt.Println(msg.Channel, msg.Data)
            }
        }
    }
    ```
* 2018-08-03 11:00

#### Version 0.6.5
* Update Command: ZRangeByScore(key string, start, stop string, isWithScores bool)([]string, error)
* 2018-07-26 18:00

#### Version 0.6.4
* New Command: ZRank(key, member string) (int, error)
* New Command: ZRangeByScore(key string, start, stop string)([]string, error)
* 2018-07-25 14:00

#### Version 0.6.3
* New Command: ZRem(key string, member... interface{})(int, error)
* 2018-07-17 13:00

#### Version 0.6.2
* New feature: Cache add Expire, used to set expire time on key
* Support RuntimeCache & RedisCache
* 2018-06-23 21:00

#### Version 0.6.1
* Bug Fixed：cache_redis issue #5 当ttl设置为0时，redis会返回“ERR invalid expire time in set”，不是forever
* 2018-05-30 09:00

#### Version 0.6
* 新增GetRedisCachePoolConf接口，用于需要设置连接池配置的场景
* 默认GetRedisCache与GetCache使用redis时，连接池设置默认使用RedisConnPool_MaxIdle, RedisConnPool_MaxActive
* 2018-05-09 18:00

#### Version 0.5
* RedisCache新增ZRange、ZRevRange、ZCard、HMGet接口
* 2018-04-25 20:00

#### Version 0.4
* RedisCache新增EVAL接口
* 2018-04-24 16:00


#### Version 0.3
* RedisCache新增ZAdd、ZCount接口
* 2018-04-04 17:00

#### Version 0.2
* RedisCache新增GetJsonObj、SetJsonObj接口
* 2018-03-14 09:00

#### Version 0.1
* 初始版本，支持runtime & redis
* 2018-03-13 18:00