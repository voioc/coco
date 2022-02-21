package cache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"github.com/voioc/coco/cache/memcached"
	credis "github.com/voioc/coco/cache/redis"
)

type cacheConfigSingle struct {
	Driver   string   `json:"driver"`
	Type     string   `json:"type"`
	Nodes    []string `json:"nodes"`
	Password string   `json:"password"`
	Expire   int32    `json:"exire"`
	Flush    int32    `json:"flush"`
}

var cacheConfig []cacheConfigSingle

var rc *redis.Client
var mem *memcache.Client

func Init() error {
	err := cacheConnect()
	if err == nil {
		// 监控配置文件是否变化
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			viper.ReadInConfig()
			cacheConnect()
		})
	}

	return err
}

// func initCache() error {
// 	// 监控配置文件是否变化
// 	viper.WatchConfig()
// 	viper.OnConfigChange(func(e fsnotify.Event) {
// 		viper.ReadInConfig()
// 		cacheConnect()
// 	})

// 	return cacheConnect()
// }

// 连接
func cacheConnect() error {
	if err := viper.UnmarshalKey("cache", &cacheConfig); err != nil {
		log.Fatalln("缓存配置格式错误")
	}

	for _, conf := range cacheConfig {
		if len(conf.Nodes) < 1 {
			fmt.Println("没有可用的节点")
			continue
		}

		if conf.Driver == "memcached" {
			mem = memcached.GetInstance(conf.Nodes)
		}

		if conf.Driver == "redis" {
			// 取第1个可用IP
			tempStr := strings.Split(conf.Nodes[0], ":") //ip:port  冒号分隔
			if len(tempStr) < 2 {
				fmt.Println("配置文件host格式有误,正确格式如 ip:port," + conf.Nodes[0])
				continue
			}

			host := tempStr[0]
			port := tempStr[1]
			rc = credis.GetInstance(host, port, conf.Password)
		}
	}

	if mem == nil && rc == nil {
		log.Fatalln("init cache failed")
	}

	return nil
}

// GetRedis 获得redis
func GetRedis() *redis.Client {
	if rc == nil {
		if err := cacheConnect(); err != nil {
			// 监控配置文件是否变化
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				viper.ReadInConfig()
				cacheConnect()
			})
		}
	}
	return rc
}

// SetCacheValue 获取缓存对象 第一个参数为key
func SetCacheValue(c context.Context, key string, value interface{}, expire int) error {
	v := ""
	if vStr, flag := value.(string); !flag {
		v, _ = jsoniter.MarshalToString(value)
	} else {
		v = vStr
	}

	return GetRedis().Set(c, key, v, time.Duration(expire)*time.Second).Err()
}

// GetCacheValue 获取缓存对象 第一个参数为key
func GetCacheValue(c context.Context, key string) (string, error) {
	return GetRedis().Get(c, key).Result()
}
