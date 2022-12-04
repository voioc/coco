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
	"strings"
	"sync"
	"time"

	"github.com/arthurkiller/rollingwriter"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
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

// func init() {
// 	// until := time.Now().Add(time.Second)
// 	// AppConfig := config.GetConfig()
// 	// for AppConfig == nil {
// 	// 	if time.Now().After(until) {
// 	// 		break
// 	// 	}

// 	// 	fmt.Println("config not init, sleep...")
// 	// 	time.Sleep(time.Second)
// 	// 	// _, err = os.Stat(filePath)
// 	// }

// 	InitDB()
// }

func InitDB() {
	dataSource := viper.GetViper().GetStringMap("db")
	// 针对每个数据库配置进行初始还
	for row := range dataSource {
		ds := DS{}
		key := fmt.Sprintf("db.%s", row) // 获取配置的key
		// 读取配置文件
		if err := viper.GetViper().UnmarshalKey(key, &ds); err != nil {
			log.Println("decode config error: ", err.Error())
			continue
		}

		// 数据库初始化
		if ds.Driver == "mysql" {
			if err := dbConnect(row, ds); err != nil {
				log.Printf("%s, db: %s\n", err.Error(), row)
				continue
			}
		}
	}

	if len(dbList) < 1 {
		log.Fatalln("Init DB failed, no aviable db.")
	}
}

// GetDB 获取指定数据库连接资源 dn->dbname
func GetDB(dn ...string) *xorm.EngineGroup {
	if len(dbList) < 1 {
		InitDB()
	}

	if len(dn) < 1 {
		return dbList["main"]
	}

	return dbList[dn[1]]
}

func dbConnect(dbName string, conf DS) error {
	lockMysql.Lock()
	defer lockMysql.Unlock()

	if _, ok := dbList[dbName]; ok {
		if dbList[dbName] != nil {
			return nil
		}
	}

	if len(conf.Dsn) < 1 {
		log.Printf("%s db dsn is empty \n", dbName)
	}

	// master := conf.Dsn[0]
	// slave := conf.Dsn[1:]

	var err error
	engine, err := xorm.NewEngineGroup("mysql", conf.Dsn)
	if err != nil {
		return err
	}

	if err := engine.Ping(); err != nil {
		return err
	}

	// engine.ShowSQL(true)
	engine.SetConnMaxLifetime(5 * time.Minute)
	engine.SetMaxIdleConns(25)
	engine.SetMaxOpenConns(50)

	path, logFile := "", ""
	if pos := strings.LastIndex(conf.Log, "/"); pos != -1 {
		path = conf.Log[0:pos]
		logFile = conf.Log[pos+1 : len(conf.Log)-4] // 去除后缀名，组件自动加.log后缀名
	}

	// env := viper.GetString("env")
	if path != "" && logFile != "" {
		config := rollingwriter.Config{
			LogPath:                path,                        //日志路径
			TimeTagFormat:          "060102150405",              //时间格式串
			FileName:               logFile,                     // 日志文件名
			MaxRemain:              3,                           // 配置日志最大存留数
			RollingPolicy:          rollingwriter.VolumeRolling, // 配置滚动策略 norolling timerolling volumerolling
			RollingTimePattern:     "* * * * * *",               // 配置时间滚动策略
			RollingVolumeSize:      "2G",                        // 配置截断文件上限大小
			WriterMode:             "none",
			BufferWriterThershould: 256,
			// Compress will compress log file with gzip
			Compress: true,
		}

		writer, err := rollingwriter.NewWriterFromConfig(&config)
		if err != nil {
			log.Println(err.Error())
		}

		// logWriter, err := os.OpenFile(conf.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		// if err != nil {
		// 	log.Println("打开数据库日志文件失败:", err.Error())
		// 	return err
		// }

		logger := xlog.NewSimpleLogger(writer)
		engine.SetLogger(logger)
	}

	engine.ShowSQL(true)

	if dbList == nil {
		dbList = make(map[string]*xorm.EngineGroup, 0)
	}

	dbList[dbName] = engine
	return nil
}
