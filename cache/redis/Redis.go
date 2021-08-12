/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-29 19:25:04
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-06-16 11:35:20
 * @FilePath: /Coco/cache/redis/Redis.go
 */
package redis

import (
	"sync"

	"github.com/go-redis/redis"
)

//使用单例模式创建redis client
var mu sync.Mutex

// GetInstance dki
func GetInstance(addr, pwd string) *redis.Client {
	mu.Lock()
	defer mu.Unlock()
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd, // no password set
		DB:       0,   // use default DB
	})
	return client
}
