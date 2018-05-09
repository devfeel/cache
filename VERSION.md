## cache版本记录：


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