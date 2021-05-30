/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-13 15:27:17
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-05-30 09:48:42
 * @FilePath: /Coco/logcus/log.go
 */
package logcus

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/voioc/coco/config"
)

var logger *logrus.Entry
var errFile *os.File

// func init() {
// 	// loggerConfig := logger.NewLogConfig()
// 	// loggerConfig.LogPath = conf.GetAppConfig().Log.Error
// 	// loggerConfig.Console = false
// 	// loggerConfig.Rotate = false
// 	// loggerConfig.Level = "DEBUG"
// 	// logger.InitLogWithConfig(loggerConfig)
// 	logrus.SetReportCaller(true)
// }

// Init 11
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

	var err error
	errlog := config.GetConfig().GetString("log.error_log")
	if errFile, err = os.OpenFile(errlog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	stdout := os.Stdout

	log := logrus.New() //实例化
	// logrus.SetReportCaller(true)
	// logger.SetLevel(logrus.DebugLevel)
	log.SetOutput(io.MultiWriter(errFile, stdout))

	//设置日志格式
	// log.SetFormatter(&logrus.TextFormatter{
	// 	TimestampFormat: "2006/01/02 15:04:05",
	// })

	log.SetFormatter(&nested.Formatter{
		// HideKeys:        true,
		TimestampFormat: "2006/01/02 15:04:05",
		FieldsOrder:     []string{"name", "age"},
	})

	_, file, line, _ := runtime.Caller(1)
	logger = log.WithFields(logrus.Fields{
		"file": file,
		"line": line,
	})
}

func GetLogger() *logrus.Entry {
	return logger
}

func OutputInfo(message ...interface{}) {
	if logger != nil {
		logger.Info(message)
	}
}

func OutputError(message ...interface{}) {
	if logger != nil {
		logger.Error(message)
	}
}

func OutputPanic(message ...interface{}) {
	if logger != nil {
		logger.Error(message)
	}
}

// // Print 记录日志
// func Print(prefix string, err ...interface{}) {

// 	if logger != nil {
// 		handle := logger.WithFields(logrus.Fields{
// 			"file": file,
// 			"line": line,
// 		})

// 		if isPrint := config.GetConfig().GetBool("log.is_print"); isPrint {
// 			Println(file, line, err)
// 		}

// 		if prefix == "info" {
// 			handle.Info(err)
// 		} else if prefix == "error" {
// 			handle.Error(err)
// 		} else if prefix == "panic" {
// 			handle.Panic(err)
// 		}
// 	}
// }

// Panic 收集panic
func RecoverPanic() {
	// lg := log.New(errlog, "[panic]: ", log.Ldate|log.Ltime|log.Llongfile)
	if info := recover(); info != nil {
		panic := debug.Stack()
		if len(panic) > 0 && logger != nil {
			logger.Panic(string(panic))
		}
	}
}
