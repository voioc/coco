/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-14 14:34:46
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-06-14 19:13:01
 * @FilePath: /Coco/db/mysql.go
 */

package db

import (
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/voioc/coco/config"
	"github.com/voioc/coco/logcus"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var engine *xorm.EngineGroup

//var onceMysql sync.Once
var lockMysql sync.Mutex

func init() {
	until := time.Now().Add(time.Second)
	AppConfig := config.GetConfig()
	for AppConfig == nil {
		if time.Now().After(until) {
			break
		}

		fmt.Println("config not init, sleep...")
		time.Sleep(time.Second)
		// _, err = os.Stat(filePath)
	}

	mysqlConn()
}

func GetMySQL() *xorm.EngineGroup {
	if engine == nil {
		mysqlConn()
	}

	return engine
}

func mysqlConn() {
	lockMysql.Lock()
	defer lockMysql.Unlock()

	// driverName := config.GetConfig().GetString("db.dsn")
	dataSourceName := config.GetConfig().GetStringSlice("db.dsn")
	if len(dataSourceName) == 0 || dataSourceName[0] == "" {
		logcus.GetLogger().Fatalln("Mysql config is empty.")
	}

	var err error
	engine, err = xorm.NewEngineGroup("mysql", dataSourceName)
	if err != nil {
		logcus.GetLogger().Fatalln("Connect mysql error: ", err.Error())
	}

	engine.ShowSQL(true)
	engine.SetConnMaxLifetime(5 * time.Minute)
	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(100)

	if config.RunEnv == "release" {
		sl := config.GetConfig().GetString("db.log")
		logWriter, err := os.OpenFile(sl, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logcus.GetLogger().Fatalln("打开数据库日志文件失败:", err)
		}

		logger := log.NewSimpleLogger(logWriter)
		logger.ShowSQL(true)
		engine.SetLogger(logger)
	}
}
