package internal

import (
	"github.com/bradfitz/gomemcache/memcache"
)

var mc *memcache.Client

type MemCachedClient struct {
	server []string
	cli    *memcache.Client
}

const (
	DefaultMemcachedMaxIdleConns = 100
	DefaultMemcachedConnTimeout  = 10
)

func GetMemcachedClient(server ...string) *memcache.Client {
	if server == nil {
		server = []string{"127.0.0.1:11211"}
	}

	cli := memcache.New(server...)
	cli.MaxIdleConns = DefaultMemcachedMaxIdleConns
	cli.Timeout = DefaultMemcachedConnTimeout
	return cli
}

func (mcc *MemCachedClient) Get(key string) (string, error) {
	item, err := mc.Get(key)
	return string(item.Value), err
}

func (mcc *MemCachedClient) Put(key string, value string, Flags uint32, Expiration int32) error {
	item := newItem(key, value, Flags, Expiration)
	err := mc.Set(item)
	return err
}

func newItem(key string, value string, Flags uint32, Expiration int32) *memcache.Item {
	return &memcache.Item{Key: key, Value: []byte(value), Flags: Flags, Expiration: Expiration}
}
