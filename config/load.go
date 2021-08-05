/*
 * @Author: Cedar
 * @Date: 2020-06-17 17:50:50
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-06-14 18:04:07
 * @FilePath: /Coco/config/load.go
 */
package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	// Version should be updated by hand at each release
	RunEnv      string
	ProjectPath string
	AppVersion  string
	GitCommit   string
	BuildTime   string
	GoVersion   string
)

var once sync.Once
var config *viper.Viper

// func init() {
// 	env := os.Getenv("RunEnv")
// 	env = strings.ToLower(env)

// 	switch env {
// 	case "dev":
// 		RunEnv = "dev"
// 	case "release":
// 		RunEnv = "release"
// 	default: // other
// 		RunEnv = "debug"
// 	}

// 	path, _ := filepath.Abs(filepath.Dir(""))
// 	config := path + "/config/config_" + RunEnv + ".toml"

// 	versionFlag := flag.Bool("v", false, "print the version")
// 	configFile := flag.String("c", config, "配置文件路径")
// 	flag.Parse()

// 	if *versionFlag {
// 		fmt.Printf("App Version: %s \n", AppVersion)
// 		fmt.Printf("Git Commit: %s \n", GitCommit)
// 		fmt.Printf("Build Time: %s \n", BuildTime)
// 		fmt.Printf("Go Version: %s \n", GoVersion)
// 		os.Exit(0)
// 	}

// 	//  configfile := build.GetConfigFile()
// 	viper.SetConfigFile(*configFile)
// 	fmt.Println("Loading config file " + *configFile)

// 	once.Do(func() {
// 		if err := viper.ReadInConfig(); err != nil {
// 			log.Fatalln("打开配置文件失败：", err)
// 		}
// 	})

// 	viper.WatchConfig()
// 	viper.OnConfigChange(func(e fsnotify.Event) {
// 		viper.ReadInConfig()
// 	})
// }

func GetConfig() *viper.Viper {
	if config == nil {
		config = viper.GetViper()
	}

	return config
}

func SetConfig(file string) {
	//  configfile := build.GetConfigFile()
	viper.SetConfigFile(file)
	fmt.Println("Loading config file " + file)

	once.Do(func() {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalln("打开配置文件失败：", err)
		}
	})
}
