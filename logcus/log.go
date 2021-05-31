/*
 * @Description: Do not edit
 * @Author: Jianxuesong
 * @Date: 2021-05-13 15:27:17
 * @LastEditors: Jianxuesong
 * @LastEditTime: 2021-05-31 19:58:43
 * @FilePath: /Coco/logcus/log.go
 */
package logcus

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/voioc/coco/config"
)

var log *logrus.Logger
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
	}

	InitLog()
}

func InitLog() *logrus.Logger {
	var err error
	errlog := config.GetConfig().GetString("log.error_log")
	if errFile, err = os.OpenFile(errlog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	log = logrus.New() //实例化
	// logrus.SetReportCaller(true)
	// logger.SetLevel(logrus.DebugLevel)
	log.SetOutput(io.MultiWriter(errFile, os.Stdout))

	//设置日志格式
	// log.SetFormatter(&logrus.TextFormatter{
	// 	TimestampFormat: "2006/01/02 15:04:05",
	// })

	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "2006/01/02 15:04:05",
		FieldsOrder:     []string{"name", "age"},
	})

	return log
}

func GetLogger() *logrus.Logger {
	return log
}

func OutputInfo(message ...interface{}) {
	if log == nil {
		InitLog()
	}

	if log != nil {
		_, file, line, _ := runtime.Caller(1)
		logger := log.WithFields(logrus.Fields{
			"file - line": fmt.Sprintf("%s:%d", file, line),
			// "line": line,
		})
		logger.Info(message)
	}
}

func OutputError(message ...interface{}) {
	if log == nil {
		InitLog()
	}

	if log != nil {
		_, file, line, _ := runtime.Caller(1)
		logger := log.WithFields(logrus.Fields{
			"file - line": fmt.Sprintf("%s:%d", file, line),
			// "line": line,
		})
		logger.Error(message)
	}
}

func OutputPanic(message ...interface{}) {
	if log != nil {
		_, file, line, _ := runtime.Caller(2)
		logger := log.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		})

		logger.Panic(message)
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
		if len(panic) > 0 && log != nil {
			log.Panic(string(panic))
		}
	}
}
