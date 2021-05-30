/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-29 19:25:04
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-05-30 16:24:31
 * @FilePath: /Coco/cache/Redis.go
 */
package redis

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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

var Client *redis.Client

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
	var cacheConfig []cacheConfigSingle
	AppConfig.UnmarshalKey("cache", &cacheConfig)

	// path := conf.RedisIniPath
	err := connect(cacheConfig)
	if err != nil {
		logcus.OutputError("初始化redis数据异常:" + err.Error())
	}
}

//使用单例模式创建redis client
var mu sync.Mutex

func getInstance(addr, pwd string) *redis.Client {
	mu.Lock()
	defer mu.Unlock()
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd, // no password set
		DB:       0,   // use default DB
	})
	return Client
}

//连接
func connect(cacheConfig []cacheConfigSingle) error {
	// redisConf, err := initConf(fileName)
	for _, conf := range cacheConfig {
		if conf.Driver == "redis" {
			if len(conf.Nodes) < 1 {
				logcus.OutputError("没有可用的redis")
				return errors.New("no useful redis address")
			}

			//取第1个可用IP
			for _, host := range conf.Nodes {
				tempStrs := strings.Split(host, ":") //ip:port  冒号分隔
				if len(tempStrs) < 2 {
					logcus.OutputError("配置文件host格式有误 ，正确格式如 ip:port ，" + host)
					continue
				}

				// weight, err := strconv.Atoi(tempStrs[2])
				// if err != nil {
				// 	return err
				// }
				// if weight < 1 {
				// 	weight = 1
				// }

				host := tempStrs[0] + ":" + tempStrs[1]
				//使用配置文件的密码
				// appConf := conf.GetAppConfig()
				Client = getInstance(host, conf.Password)
				// lastConfigs = *redisConf
				break
			}
		}
	}

	// go watchConfig(fileName)

	return nil
}
