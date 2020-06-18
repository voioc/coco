package lib

import (
	"log"
	"os"
	"runtime"
	"runtime/debug"

	"../config"

	"github.com/voioc/logrus"
)

var logger *logrus.Logger
var errFile *os.File

// Init 11
func Init() {
	var err error
	errlog := config.GetConfig().GetString("log.error")
	if errFile, err = os.OpenFile(errlog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	// level := service.GetConfig().GetString("log.level")
	level := "debug"

	logger = logrus.New() //实例化
	logger.Out = errFile  //设置输出

	// 设置日志级别
	logger.SetLevel(logrus.ErrorLevel)
	if level == "info" {
		logger.SetLevel(logrus.InfoLevel)
	}

	//设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
	})

	// Logger = log.New(errorFile, "[Info]", log.Ldate|log.Ltime)
}

// Print 记录日志
func Print(prefix string, err ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	// Logger.SetPrefix("[" + prefix + "]")
	// Logger.Println(file, line, err)

	if logger != nil {
		handle := logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		})

		// if isPrint := service.GetConfig().GetBool("log.is_print"); isPrint {
		// 	Println(file, line, err)
		// }

		if prefix == "info" {
			handle.Info(err)
		} else if prefix == "error" {
			handle.Error(err)
		} else if prefix == "panic" {
			handle.Panic(err)
		}
	}

}

// Panic 收集panic
func Panic() {
	// lg := log.New(errlog, "[panic]: ", log.Ldate|log.Ltime|log.Llongfile)
	if info := recover(); info != nil {
		panic := debug.Stack()

		if len(panic) > 0 {
			_, file, line, _ := runtime.Caller(2)

			if logger != nil {
				logger.WithFields(logrus.Fields{
					"file": file,
					"line": line,
				}).Panic(string(panic[:]))
			}
			// Logger.SetPrefix("[panic] ")
			// Logger.Println(file, line, string(panic[:]))

			// WriteLog("panic", string(panic[:]))
		}
	}
}

// Fatalln 继承
func Fatalln(v ...interface{}) {
	log.Fatalln(v)
}

// Println 继承
func Println(v ...interface{}) {
	log.Println(v)
}

// Printf 继承
func Printf(format string, v ...interface{}) {
	log.Printf(format, v)
}

// Defer 函数
func Defer() {
	errFile.Close()
}
