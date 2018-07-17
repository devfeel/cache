## cache版本记录：

#### Version 0.6.2
* New Command: ZRem(key string, member... interface{})(int, error)
* 2018-07-17 13:00

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