package redis

import (
	"fmt"
	"testing"
)

var rc *redisCache

func init() {
	rc = NewRedisCache("127.0.0.1:6379")
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
