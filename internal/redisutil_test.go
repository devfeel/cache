package internal

import (
	"testing"
	"fmt"
	"reflect"
)

var rc *RedisClient
func init() {
	rc = GetRedisClient("redis://192.168.8.175:6379/0", 10, 10)
}

func TestRedisClient_Ping(t *testing.T) {
	res, err :=rc.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res, reflect.TypeOf(res))
}

func TestRedisClient_Del(t *testing.T) {
	res, err :=rc.Del("1", "", "f12")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res, reflect.TypeOf(res))
}

func TestRedisClient_HDel(t *testing.T) {
	res, err :=rc.HDel("h1", "f11", "f12")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res, reflect.TypeOf(res))
}

func TestRedisClient_HSetNX(t *testing.T) {
	res, err :=rc.HSetNX("h1", "f11", "f12")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res, reflect.TypeOf(res))
}

func TestRedisClient_HVals(t *testing.T) {
	res, err :=rc.HVals("h1")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res, reflect.TypeOf(res))
}

func TestRedisClient_SDiff(t *testing.T) {
	res, err :=rc.SDiff("skey1", "skey2")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res, reflect.TypeOf(res))
}