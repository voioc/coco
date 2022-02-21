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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

//使用单例模式创建redis client
var mu sync.Mutex

// GetInstance dki
func GetInstance(host, port, pwd string) *redis.Client {
	mu.Lock()
	defer mu.Unlock()

	addr := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:               addr,
		Password:           pwd, // no password set
		DB:                 0,   // use default DB
		PoolSize:           500,
		IdleTimeout:        time.Second,
		IdleCheckFrequency: 10 * time.Second,
		MinIdleConns:       3,
		MaxRetries:         3, // 最大重试次数
		DialTimeout:        2 * time.Second,
	})

	ctx := context.Background()
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("contect redis error: %s, %s\n", pong, err.Error())
		return nil
	}

	fmt.Println("redis address: "+addr, pong)
	return client
}
