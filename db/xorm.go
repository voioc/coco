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
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/voioc/coco/logcus"
	"github.com/voioc/coco/logzap"
	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
)

// DBStruct 数据库结构
type DS struct {
	Driver string
	Dsn    []string
	Log    string
}

var dbList map[string]*xorm.EngineGroup

// var engine *xorm.EngineGroup
var lockMysql sync.Mutex

func init() {
	// until := time.Now().Add(time.Second)
	// AppConfig := config.GetConfig()
	// for AppConfig == nil {
	// 	if time.Now().After(until) {
	// 		break
	// 	}

	// 	fmt.Println("config not init, sleep...")
	// 	time.Sleep(time.Second)
	// 	// _, err = os.Stat(filePath)
	// }

	InitDB()
}

func InitDB() {
	dataSource := viper.GetViper().GetStringMap("db")
	// 针对每个数据库配置进行初始还
	for row := range dataSource {
		ds := DS{}
		key := fmt.Sprintf("db.%s", row) // 获取配置的key
		// 读取配置文件
		if err := viper.GetViper().UnmarshalKey(key, &ds); err != nil {
			fmt.Println("decode config error: ", err.Error())
			continue
		}

		// 数据库初始化
		if ds.Driver == "mysql" {
			if err := mysqlConnect(row, ds); err != nil {
				fmt.Printf("%s, db: %s\n", err.Error(), row)
				continue
			}
		}
	}

	if len(dbList) < 1 {
		log.Fatalln("Init DB failed, no aviable db.")
	}
}

// GetDB 获取指定数据库连接资源 dn->dbname
func GetMySQL(dn ...string) *xorm.EngineGroup {
	if len(dbList) < 1 {
		InitDB()
	}

	if len(dn) < 1 {
		return dbList["main"]
	}

	return dbList[dn[1]]
}

func mysqlConnect(dbName string, conf DS) error {
	lockMysql.Lock()
	defer lockMysql.Unlock()

	if _, ok := dbList[dbName]; ok {
		if dbList[dbName] != nil {
			return nil
		}
	}

	if len(conf.Dsn) < 1 {
		fmt.Printf("%s db dsn is empty", dbName)
	}

	// master := conf.Dsn[0]
	// slave := conf.Dsn[1:]

	var err error
	engine, err := xorm.NewEngineGroup("mysql", conf.Dsn)
	if err != nil {
		logcus.GetLogger().Fatalln("Connect mysql error: ", err.Error())
	}

	engine.ShowSQL(true)
	engine.SetConnMaxLifetime(5 * time.Minute)
	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(100)

	env := viper.GetString("env")
	if env == "release" {
		logWriter, err := os.OpenFile(conf.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logzap.E("打开数据库日志文件失败:", err)
		}

		logger := xlog.NewSimpleLogger(logWriter)
		logger.ShowSQL(true)
		engine.SetLogger(logger)
	}

	if dbList == nil {
		dbList = make(map[string]*xorm.EngineGroup, 0)
	}

	dbList[dbName] = engine
	return nil
}
