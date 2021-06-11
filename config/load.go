/*
 * @Author: Cedar
 * @Date: 2020-06-17 17:50:50
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-06-11 18:23:02
 * @FilePath: /Coco/config/load.go
 */
package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	// Version should be updated by hand at each release
	RunEnv      string
	ProjectPath string
	Version     string
	GitCommit   string
	BuildTime   string
	GoVersion   string
)

var once sync.Once
var config *viper.Viper

func init() {
	env := os.Getenv("RunEnv")
	env = strings.ToLower(env)

	switch env {
	case "test":
		RunEnv = "test"
	case "release":
		RunEnv = "release"
	default: //dev and other
		RunEnv = "dev"
	}

	path, _ := filepath.Abs(filepath.Dir(""))
	// config := path[0:strings.LastIndex(path, "/")] + "/config/config_debug.json"
	config := path + "/config/config_" + RunEnv + ".toml"

	versionFlag := flag.Bool("v", false, "print the version")
	configFile := flag.String("c", config, "配置文件路径")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("App Version: %s \n", Version)
		fmt.Printf("Git Commit: %s \n", GitCommit)
		fmt.Printf("Build Time: %s \n", BuildTime)
		fmt.Printf("Go Version: %s \n", GoVersion)
		os.Exit(0)
	}

	//  configfile := build.GetConfigFile()
	viper.SetConfigFile(*configFile)
	fmt.Println("Loading config file " + *configFile)

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
