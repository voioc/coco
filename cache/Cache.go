package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/voioc/coco/cache/memcached"
	credis "github.com/voioc/coco/cache/redis"
	"github.com/voioc/coco/config"
	"github.com/voioc/coco/logcus"
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

func init() {
	until := time.Now().Add(5 * time.Second)
	AppConfig := config.GetConfig()
	for AppConfig == nil {
		if time.Now().After(until) {
			break
		}

		fmt.Println("config not init, sleep...")
		time.Sleep(time.Second)
		// _, err = os.Stat(filePath)
	}

	//必须先调用这个函数初始化SDk  (或者在init函数初始化)
	AppConfig.UnmarshalKey("cache", &cacheConfig)
	err := connect()
	if err != nil {
		logcus.Error("初始化redis数据异常:" + err.Error())
	}
}

// 连接
func connect() error {
	for _, conf := range cacheConfig {
		if len(conf.Nodes) < 1 {
			logcus.Error("没有可用的节点")
			return errors.New("no useful address")
		}

		if conf.Driver == "memcached" {
			var servers []string
			for _, row := range conf.Nodes {
				servers = append(servers, row)
			}
			mem = memcached.GetInstance(servers)
		}

		if conf.Driver == "redis" {
			//取第1个可用IP
			for _, host := range conf.Nodes {
				tempStr := strings.Split(host, ":") //ip:port  冒号分隔
				if len(tempStr) < 2 {
					logcus.Error("配置文件host格式有误 ，正确格式如 ip:port ，" + host)
					continue
				}

				host := tempStr[0] + ":" + tempStr[1]

				//使用配置文件的密码
				rc = credis.GetInstance(host, conf.Password)
				break
			}
		}
	}

	return nil
}

// GetRedis 获得redis
func GetRedis() *redis.Client {
	if rc == nil {
		connect()
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
