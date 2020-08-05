package cache

import (
	"fmt"
	// log "lemon/lib/log"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	GoCache "github.com/patrickmn/go-cache"
	"github.com/voioc/coco/config"
	log "github.com/voioc/coco/log"
)

type CacheConfigSingle struct {
	Driver   string   `json:"driver"`
	Nodes    []string `json:"nodes"`
	Password []string `json:"password"`
	Expire   int32    `json:"exire"`
	Flush    int32    `json:"flush"`
}

var cacheConfig []CacheConfigSingle

var cacheClient *GoCache.Cache
var memClient *memcache.Client
var redisClient *redis.Client
var redisClusterClient *redis.ClusterClient

func Init() {
	config.GetConfig().UnmarshalKey("cache", &cacheConfig)

	for _, cache := range cacheConfig {
		if cache.Driver == "cache" {
			cacheClient = GoCache.New(time.Duration(cache.Expire)*time.Second, time.Duration(cache.Flush)*time.Second)
		}

		if cache.Driver == "memcached" {
			var servers []string
			for _, row := range cache.Nodes {
				servers = append(servers, row)
			}
			// memcache 文件内
			memClient = memcache.New(servers[0:]...)
		}

		if cache.Driver == "redis" {
			password := ""
			if len(cache.Nodes) == len(cache.Password) {
				password = cache.Password[0]
			}

			if len(cache.Nodes) == 1 {
				redisClient = redis.NewClient(&redis.Options{
					Addr:     cache.Nodes[0],
					Password: password, // no password set
					DB:       0,        // use default DB
				})

				if ping, err := redisClient.Ping().Result(); err != nil {
					log.Print("info", "Test Redis Server:"+ping+"the error is: "+err.Error())
				}
			} else {
				redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
					Addrs:    cache.Nodes,
					Password: password,
				})

				if ping, err := redisClusterClient.Ping().Result(); err != nil {
					log.Print("info", "Test Redis Server:"+ping+"the error is: "+err.Error())
				}
			}

			// err := redisClient.Do("SET", "cache", cacheConfig, "EX", "300")
			// fmt.Println(err)
		}
	}
}

// GetRedis 获得redis
func GetRedis() interface{} {
	if redisClient == nil {
		return redisClusterClient
	}
	return redisClient
}

// GetRedisCluster 获得redisCluster
func GetRedisCluster() *redis.ClusterClient {
	return redisClusterClient
}

// GetCache sdfsdf
func GetCache(cacheKey string, data interface{}) (bool, error) {

	isGet, cacheErr := getCacheByDriver(cacheKey, cacheConfig[0].Driver, data)

	if isGet == false || cacheErr != nil {
		isGet, cacheErr = getCacheByDriver(cacheKey, cacheConfig[1].Driver, data)

		// 从二级缓存拿到数据的话写入一级缓存
		if isGet == true && cacheErr == nil {
			setCacheByDriver(cacheKey, cacheConfig[0].Driver, data, 1800)
		}
	}

	return isGet, cacheErr
}

// getCacheByDriver 根据不同的缓存驱动获取数据
func getCacheByDriver(cacheKey, driver string, dataStruct interface{}) (bool, error) {
	var CacheGet bool = false
	var value string = ""
	var err error
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch driver {
	case "memcached":
		valueTmp, err := memClient.Get(cacheKey)
		if err != nil {
			log.Print("info", "Cache Memcached get cache, the key is: "+cacheKey+" the error is: "+err.Error())
			return CacheGet, fmt.Errorf("[info] Cache Memcached get cache: %s", err.Error())
		}

		if len(valueTmp.Value) == 0 {
			return false, nil
		}
		value = string(valueTmp.Value)

	case "redis":
		if redisClusterClient == nil {
			value, err = redisClient.Do("GET", cacheKey).String()
		} else {
			value, err = redisClusterClient.Do("GET", cacheKey).String()
		}

		if err != nil {
			log.Print("info", "Cache there is error when get value from key: "+cacheKey+" error is: "+err.Error())
			return false, fmt.Errorf("[info] Cache there is error when get value from key %s : %s", cacheKey, err.Error())
		}
	}

	if value == "" {
		return false, nil
	}

	if err := json.UnmarshalFromString(value, dataStruct); err != nil {
		log.Print("info", "Cache get cache the key is: "+cacheKey+" the value is: "+value+" the error is: "+err.Error())
		return false, fmt.Errorf("[info]LCache Redisc get cache: %s", err.Error())
	}

	CacheGet = true

	return CacheGet, nil
}

// SetCache 写入缓存
func SetCache(cacheKey string, data interface{}, expire int32) error {
	setCacheByDriver(cacheKey, cacheConfig[0].Driver, data, expire)
	setCacheByDriver(cacheKey, cacheConfig[1].Driver, data, expire)

	return nil
}

// SetCacheByDriver 设置缓存
func setCacheByDriver(cacheKey, driver string, data interface{}, expire int32) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	dataStr, err := json.MarshalToString(data)
	if err != nil {
		log.Print("info", "Cache marshall struct: "+err.Error())
		return fmt.Errorf("[info] Cache marshall struct: %s", err.Error())
	}

	switch driver {
	case "memcached":
		item := &memcache.Item{
			Key:        cacheKey,
			Value:      []byte(dataStr),
			Expiration: expire,
		}

		err := memClient.Set(item)
		if err != nil {
			log.Print("info", "Cache Memcached set cache: "+err.Error())
			return fmt.Errorf("[info]Cache Memcached set cache: %s", err.Error())
		}

	case "redis":
		var errCmd *redis.Cmd
		if redisClusterClient == nil {
			errCmd = redisClient.Do("SET", cacheKey, dataStr, "EX", expire)
		} else {
			errCmd = redisClusterClient.Do("SET", cacheKey, dataStr, "EX", expire)
		}

		if errCmd.Err() != nil {
			log.Print("info", "Cache Redis set cache:: "+errCmd.Err().Error())
			return fmt.Errorf("[info] Cache Redis set cache: %s", errCmd.Err().Error())
		}
	}

	return nil
}
