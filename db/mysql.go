/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-14 14:34:46
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-05-30 16:22:40
 * @FilePath: /Coco/db/mysql.go
 */

package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/voioc/coco/config"
	"github.com/voioc/coco/logcus"
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
	dataSourceName := config.GetConfig().GetString("db.mysql.dsn")
	if dataSourceName == "" {
		logcus.OutputError("Mysql config is empty.")
		os.Exit(504)
	}

	engine, err := xorm.NewEngineGroup("mysql", dataSourceName)
	if err != nil {
		logcus.OutputError(fmt.Sprintf("Connect mysql error: %s", err.Error()))
		os.Exit(504)
	}

	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(100)

	sl := config.GetConfig().GetString("db.mysql.sql_log")
	sqlLog, err := os.OpenFile(sl, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	} else {
		engine.ShowSQL(true)
		engine.SetLogger(xorm.NewSimpleLogger(sqlLog))
	}

}
