package memcached

import "github.com/devfeel/cache/internal"

type memcachedCache struct {
	server []string
}

func init() {

}

func NewMemcachedCache (servers ...string) *memcachedCache{
	cache := &memcachedCache{server:servers}
	return cache
}

func (mc *memcachedCache) Get(key string) (interface{}, error) {
	client := internal.GetMemcachedClient(mc.server...)
	item, err := client.Get(key)
	return string(item.Value), err
}

// Exist return true if value cached by given key
func (mc *memcachedCache) Exists(key string) (bool, error){
	return false, nil
}

// GetString returns value string format by given key
func (mc *memcachedCache) GetString(key string) (string, error) {
	return "", nil
}
// GetInt returns value int format by given key
func (mc *memcachedCache) GetInt(key string) (int, error) {
	return 0, nil
}
// GetInt64 returns value int64 format by given key
func (mc *memcachedCache) GetInt64(key string) (int64, error) {
	return 0, nil
}
// Set cache value by given key
func (mc *memcachedCache) Set(key string, v interface{}, ttl int64) error {
	return nil
}
// Incr increases int64-type value by given key as a counter
// if key not exist, before increase set value with zero
func (mc *memcachedCache) Incr(key string) (int64, error) {
	return 0, nil
}
// Decr decreases int64-type value by given key as a counter
// if key not exist, before increase set value with zero
func (mc *memcachedCache) Decr(key string) (int64, error) {
	return 0, nil
}
// Delete delete cache item by given key
func (mc *memcachedCache) Delete(key string) error {
	return nil
}
// ClearAll clear all cache items
func (mc *memcachedCache) ClearAll() error {
	return nil
}