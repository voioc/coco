/*
 * @Author: Cedar
 * @Date: 2020-06-17 17:50:50
 * @LastEditors: Cedar
 * @LastEditTime: 2020-06-17 17:52:04
 * @FilePath: /Coco/config/load.go
 */
package config

import (
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	// "gopkg.in/fsnotify.v1"
)

var once sync.Once
var config *viper.Viper

func LoadConfig(configfile *string) {
	//  configfile := build.GetConfigFile()
	viper.SetConfigFile(*configfile)

	once.Do(func() {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalln("打开配置文件失败：", err)
		}
	})

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()
	})
}

func GetConfig() *viper.Viper {
	if config == nil {
		config = viper.GetViper()
	}

	return config
}
