package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	// DefaultGCInterval means gc interval.
	DefaultGCInterval       = 60 * time.Second // 1 minute
	ZeroInt64         int64 = 0
)

// RuntimeItem store runtime cache item.
type RuntimeItem struct {
	value      interface{}
	createTime time.Time
	ttl        time.Duration
}

//check item is expire
func (mi *RuntimeItem) isExpire() bool {
	// 0 means forever
	if mi.ttl == 0 {
		return false
	}
	return time.Now().Sub(mi.createTime) > mi.ttl
}

// RuntimeCache is runtime cache adapter.
// it contains a RW locker for safe map storage.
type RuntimeCache struct {
	sync.RWMutex
	gcInterval time.Duration
	items      map[string]*RuntimeItem
}

// NewRuntimeCache returns a new *RuntimeCache.
func NewRuntimeCache() *RuntimeCache {
	cache := &RuntimeCache{items: make(map[string]*RuntimeItem), gcInterval: DefaultGCInterval}
	go cache.gc()
	return cache
}

// Get cache from runtime cache.
// if non-existed or expired, return nil.
func (ca *RuntimeCache) Get(key string) (interface{}, error) {
	ca.RLock()
	defer ca.RUnlock()
	if item, ok := ca.items[key]; ok {
		if item.isExpire() {
			return nil, nil
		}
		return item.value, nil
	}
	return nil, nil
}

// returns value string format by given key
// if non-existed or expired, return "".
func (ca *RuntimeCache) GetString(key string) (string, error) {
	v, err := ca.Get(key)
	if err != nil || v == nil {
		return "", nil
	} else {
		return fmt.Sprint(v), nil
	}
}

// returns value int format by given key
// if non-existed or expired, return 0.
func (ca *RuntimeCache) GetInt(key string) (int, error) {
	v, err := ca.GetString(key)
	if err != nil || v == "" {
		return 0, nil
	} else {
		i, e := strconv.Atoi(v)
		if e != nil {
			return 0, nil
		} else {
			return i, nil
		}
	}
}

// returns value int64 format by given key
// if non-existed or expired, return 0.
func (ca *RuntimeCache) GetInt64(key string) (int64, error) {
	v, err := ca.GetString(key)
	if err != nil || v == "" {
		return ZeroInt64, nil
	} else {
		i, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			return ZeroInt64, nil
		} else {
			return i, nil
		}
	}
}

// Set cache to runtime.
// ttl is second, if ttl is 0, it will be forever till restart.
func (ca *RuntimeCache) Set(key string, value interface{}, ttl int64) error {
	ca.Lock()
	defer ca.Unlock()
	ca.items[key] = &RuntimeItem{
		value:      value,
		createTime: time.Now(),
		ttl:        time.Duration(ttl) * time.Second,
	}
	return nil
}

// Incr increase int64 counter in runtime cache.
func (ca *RuntimeCache) Incr(key string) (int64, error) {
	ca.RLock()
	item, ok := ca.items[key]
	ca.RUnlock()
	if !ok {
		//if not exists, auto set new with 0
		ca.Set(key, ZeroInt64, 0)
		//reload
		ca.RLock()
		item, _ = ca.items[key]
		ca.RUnlock()
	}

	switch item.value.(type) {
	case int:
		item.value = item.value.(int) + 1
	case int32:
		item.value = item.value.(int32) + 1
	case int64:
		item.value = item.value.(int64) + 1
	case uint:
		item.value = item.value.(uint) + 1
	case uint32:
		item.value = item.value.(uint32) + 1
	case uint64:
		item.value = item.value.(uint64) + 1
	default:
		return 0, errors.New("item val is not (u)int (u)int32 (u)int64")
	}

	val, _ := strconv.ParseInt(fmt.Sprint(item.value), 10, 64)
	return val, nil
}

// Decr decrease counter in runtime cache.
func (ca *RuntimeCache) Decr(key string) (int64, error) {
	ca.RLock()
	item, ok := ca.items[key]
	ca.RUnlock()
	if !ok {
		//if not exists, auto set new with 0
		ca.Set(key, ZeroInt64, 0)
		//reload
		ca.RLock()
		item, _ = ca.items[key]
		ca.RUnlock()
	}
	switch item.value.(type) {
	case int:
		item.value = item.value.(int) - 1
	case int64:
		item.value = item.value.(int64) - 1
	case int32:
		item.value = item.value.(int32) - 1
	case uint:
		if item.value.(uint) > 0 {
			item.value = item.value.(uint) - 1
		} else {
			return 0, errors.New("item val is less than 0")
		}
	case uint32:
		if item.value.(uint32) > 0 {
			item.value = item.value.(uint32) - 1
		} else {
			return 0, errors.New("item val is less than 0")
		}
	case uint64:
		if item.value.(uint64) > 0 {
			item.value = item.value.(uint64) - 1
		} else {
			return 0, errors.New("item val is less than 0")
		}
	default:
		return 0, errors.New("item val is not int int64 int32")
	}
	val, _ := strconv.ParseInt(fmt.Sprint(item.value), 10, 64)
	return val, nil
}

// Exist check item exist in runtime cache.
func (ca *RuntimeCache) Exists(key string) (bool, error) {
	ca.RLock()
	defer ca.RUnlock()
	if v, ok := ca.items[key]; ok {
		return !v.isExpire(), nil
	}
	return false, nil
}

// Delete item in runtime cacha.
// if not exists, we think it's success
func (ca *RuntimeCache) Delete(key string) error {
	ca.Lock()
	defer ca.Unlock()
	if _, ok := ca.items[key]; !ok {
		//if not exists, we think it's success
		return nil
	}
	delete(ca.items, key)
	if _, ok := ca.items[key]; ok {
		return errors.New("delete key error")
	}
	return nil
}

// Expire Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
// timeout time duration is second
// if not exists key, return 0, nil
func (ca *RuntimeCache) Expire(key string, timeOutSeconds int) (int, error){
	item, ok := ca.getRuntimeItem(key)
	if !ok {
		return 0, nil
	}else{
		item.ttl = time.Duration(timeOutSeconds) * time.Second
		ca.Lock()
		ca.items[key] = item
		ca.Unlock()
		return timeOutSeconds, nil
	}
}

// ClearAll will delete all item in runtime cache.
func (ca *RuntimeCache) ClearAll() error {
	ca.Lock()
	defer ca.Unlock()
	ca.items = make(map[string]*RuntimeItem)
	return nil
}

func (ca *RuntimeCache) gc() {
	for {
		<-time.After(ca.gcInterval)
		if ca.items == nil {
			return
		}
		for name := range ca.items {
			ca.itemExpired(name)
		}
	}
}

// getRuntimeItem get RuntimeItem by key
func (ca *RuntimeCache) getRuntimeItem(key string) (*RuntimeItem, bool){
	ca.RLock()
	defer ca.RUnlock()
	item, ok := ca.items[key]
	return item, ok
}

// itemExpired returns true if an item is expired.
func (ca *RuntimeCache) itemExpired(key string) bool {
	itm, ok := ca.getRuntimeItem(key)
	if !ok {
		return true
	}
	if itm.isExpire() {
		ca.Lock()
		delete(ca.items, key)
		ca.Unlock()
		return true
	}
	return false
}
