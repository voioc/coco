package memcached

import (
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
)

//使用单例模式创建redis client
var mu sync.Mutex

// var client *memcache.Client

// GetInstance dki
func GetInstance(servers []string) *memcache.Client {
	mu.Lock()
	defer mu.Unlock()
	client := memcache.New(servers[0:]...)
	return client
}
