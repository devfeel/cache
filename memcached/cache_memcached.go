package memcached

import "github.com/devfeel/cache/internal"

type memcachedCache struct {
	server []string
}

func init() {

}

func NewMemcachedCache (servers []string) *memcachedCache{
	cache := &memcachedCache{server:servers}
	return cache
}

func (mc *memcachedCache) Get(key string) (string, error) {
	client := internal.GetMemcachedClient(mc.server...)
	client.Get(key)
}