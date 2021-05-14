package cache

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
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

var Redis *redis.Client

// var cacheClient *GoCache.Cache
var Mem *memcache.Client

// Init 初始化
func Init() {
	CacheConn("")
}

func CacheConn(driver string) {
	var cacheConfig []cacheConfigSingle
	config.GetConfig().UnmarshalKey("cache", &cacheConfig)

	for _, cache := range cacheConfig {
		if cache.Driver == "memcached" {
			if driver == "" || driver == cache.Driver {
				var servers []string
				for _, row := range cache.Nodes {
					servers = append(servers, row)
				}
				// memcache 文件内
				Mem = memcache.New(servers[0:]...)
			}
		}

		if cache.Driver == "redis" {
			if driver == "" || driver == cache.Driver {
				password := ""
				if len(cache.Nodes) == len(cache.Password) {
					password = cache.Password
				}

				Redis = redis.NewClient(&redis.Options{
					Addr:     cache.Nodes[0],
					Password: password, // no password set
					DB:       0,        // use default DB
				})

				if ping, err := Redis.Ping().Result(); err != nil {
					logcus.OutputError("info", "Test Redis Server:"+ping+"the error is: "+err.Error())
				}
			}
		}
	}
}

// GetRedis 获得redis
func GetRedis() *redis.Client {
	if Redis == nil {
		CacheConn("redis")
	}

	return Redis
}

// // GetCache 获取缓存
// func GetCache(cacheKey string, data interface{}) (bool, error) {
// 	isGet, cacheErr := getCacheByDriver(cacheKey, cacheConfig[0].Driver, data)
// 	if (isGet == false || cacheErr != nil) && len(cacheConfig) > 1 {
// 		isGet, cacheErr = getCacheByDriver(cacheKey, cacheConfig[1].Driver, data)

// 		// 从二级缓存拿到数据的话写入一级缓存
// 		if isGet == true && cacheErr == nil {
// 			setCacheByDriver(cacheKey, cacheConfig[0].Driver, data, 1800)
// 		}
// 	}

// 	return isGet, cacheErr
// }

// // getCacheByDriver 根据不同的缓存驱动获取数据
// func getCacheByDriver(cacheKey, driver string, dataStruct interface{}) (bool, error) {
// 	var CacheGet bool = false
// 	var value string = ""
// 	var err error
// 	var json = jsoniter.ConfigCompatibleWithStandardLibrary

// 	switch driver {
// 	case "memcached":
// 		valueTmp, err := memClient.Get(cacheKey)
// 		if err != nil {
// 			info := fmt.Sprintf("Cache Memcached error | {key} %s {error} %s ", cacheKey, err.Error())
// 			logcus.Print("info", info)
// 			return CacheGet, fmt.Errorf(info)
// 		}

// 		if len(valueTmp.Value) == 0 {
// 			return false, nil
// 		}
// 		value = string(valueTmp.Value)

// 	case "redis":
// 		if value, err = redisClient.Do("GET", cacheKey).String(); err != nil {
// 			info := fmt.Sprintf("Cache Reids error | {key} %s {error} %s ", cacheKey, err.Error())
// 			logcus.Print("info", info)
// 			return false, fmt.Errorf(info)
// 		}

// 		// if err != nil {
// 		// 	logcus.Print("info", "[Cache] there is error when get value from key: "+cacheKey+" error is: "+err.Error())
// 		// 	return false, fmt.Errorf("[info] Cache there is error when get value from key %s : %s", cacheKey, err.Error())
// 		// }
// 	}

// 	if value != "" {
// 		if err := json.UnmarshalFromString(value, dataStruct); err != nil {
// 			info := fmt.Sprintf("Cache Format error | {key} %s {value} %s {error} %s ", cacheKey, value, err.Error())
// 			logcus.Print("info", info)
// 			return false, fmt.Errorf(info)
// 		}
// 	}

// 	// CacheGet = true

// 	return true, nil
// }

// // SetCache 写入缓存
// func SetCache(cacheKey string, data interface{}, expire int32) error {
// 	setCacheByDriver(cacheKey, cacheConfig[0].Driver, data, expire)
// 	if len(cacheConfig) > 1 {
// 		setCacheByDriver(cacheKey, cacheConfig[1].Driver, data, expire)
// 	}

// 	return nil
// }

// // SetCacheByDriver 设置缓存
// func setCacheByDriver(cacheKey, driver string, data interface{}, expire int32) error {
// 	var json = jsoniter.ConfigCompatibleWithStandardLibrary

// 	dataStr, err := json.MarshalToString(data)
// 	if err != nil {
// 		logcus.Print("info", "Cache marshall struct: "+err.Error())
// 		return fmt.Errorf("[info] Cache marshall struct: %s", err.Error())
// 	}

// 	switch driver {
// 	case "memcached":
// 		if memClient == nil {
// 			conn()
// 		}

// 		item := &memcache.Item{
// 			Key:        cacheKey,
// 			Value:      []byte(dataStr),
// 			Expiration: expire,
// 		}

// 		err := memClient.Set(item)
// 		if err != nil {
// 			logcus.Print("info", "Cache Memcached set cache: "+err.Error())
// 			return fmt.Errorf("[info]Cache Memcached set cache: %s", err.Error())
// 		}

// 	case "redis":
// 		if redisClient == nil {
// 			conn()
// 		}

// 		if expire == -1 {
// 			err = redisClient.Do("SET", cacheKey, dataStr).Err()
// 		} else {
// 			err = redisClient.Do("SET", cacheKey, dataStr, "EX", expire).Err()
// 		}

// 		if err != nil {
// 			logcus.Print("info", "Cache Redis set cache:: "+err.Error())
// 			return fmt.Errorf("[info] Cache Redis set cache: %s", err.Error())
// 		}
// 	}

// 	return nil
// }
