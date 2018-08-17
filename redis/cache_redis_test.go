package redis

import (
	"fmt"
	"testing"
)

var rc *redisCache

func init() {
	rc = NewRedisCache("redis://192.168.8.175:6379/0", 20, 20)
}

func TestRedisCache_Publish(t *testing.T) {
	fmt.Println(rc.Publish("channel-test", "test message"))
}

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

func TestRedisCache_Ping(t *testing.T) {
	fmt.Println(rc.Ping)
}

func TestRedisCache_ZAdd(t *testing.T){
	fmt.Println(rc.ZAdd("dottest", 1, 1))
}

func TestRedisCache_Set(t *testing.T) {
	fmt.Println(rc.Set("dottest", 1, 0))
}

func TestRedisCache_ZCount(t *testing.T){
	rc.ZAdd("dottest", 1, 1)
	rc.ZAdd("dottest", 2, 2)
	rc.ZAdd("dottest", 3, 3)
	rc.ZAdd("dottest", 10,10)
	rc.ZAdd("dottest", 11, 11)
	fmt.Println(rc.ZCount("dottest", 1, 10))
}

func TestRedisCache_HDel(t *testing.T) {
	fmt.Println(rc.HDel("h1", "f11", "hf1"))
}

func TestRedisCache_HGet(t *testing.T) {
	fmt.Println(rc.HGet("h1", "hf2"))
}

func TestRedisCache_HExists(t *testing.T) {
	fmt.Println(rc.HExists("hkey1", "hkey1field1"))
}

func TestRedisCache_HIncrBy(t *testing.T) {
	fmt.Println(rc.HIncrBy("hkey1", "hint1", 1))
}
func TestRedisCache_HIncrByFloat(t *testing.T) {
	fmt.Println(rc.HIncrByFloat("hkey1","hfloat1", 0.01))
}

func TestRedisCache_HKeys(t *testing.T) {
	fmt.Println(rc.HKeys("hkey1"))
}

func TestRedisCache_HLen(t *testing.T) {
	fmt.Println(rc.HLen("hkey1"))
}

func TestRedisCache_HVals(t *testing.T) {
	fmt.Println(rc.HVals("hkey1"))
}

func TestRedisCache_HSet(t *testing.T) {
	fmt.Println(rc.HSet("hkey1", "hkey1field1", "hkey1field1value"))
	fmt.Println(rc.HSet("hkey1", "hkey1field2", "hkey1field2value"))
}

func TestRedisCache_HSetNX(t *testing.T) {
	fmt.Println(rc.HSetNX("hkey1", "hkey1field1", "hkey1field1value"))
}
