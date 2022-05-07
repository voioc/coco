package logzap

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// var logzap = zap.New(initCore(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.WarnLevel), zap.AddCaller())
var logzap *zap.Logger

type ContextKey string

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

func Zap() *zap.Logger {
	return logzap
}

func InitZap() {
	logzap = zap.New(initCore(), zap.AddCallerSkip(1), zap.AddCaller())
}

func initCore() zapcore.Core {
	logPath := "runtime/app.log"
	if errlog := viper.GetString("log.error"); errlog != "" {
		logPath = errlog
	}

	if closed := viper.GetBool("log.closed"); closed {
		logPath = "/dev/null"
	}

	maxSize := viper.GetInt("log.max_size")
	if maxSize == 0 {
		maxSize = 5120
	}

	maxAge := viper.GetInt("log.max_age")
	if maxSize == 0 {
		maxAge = 7
	}

	maxBackup := viper.GetInt("log.max_backup")
	if maxBackup == 0 {
		maxBackup = 5
	}

	opts := []zapcore.WriteSyncer{
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath, // fmt.Sprintf("%s%s/%s.log", logPath, name, name), // ⽇志⽂件路径
			MaxSize:    maxSize, // 单位为MB,默认为100MB
			MaxAge:     maxAge,  // 文件最多保存多少天
			MaxBackups: maxBackup,
			LocalTime:  true,  // 采用本地时间
			Compress:   false, // 是否压缩日志
		}),
	}

	// if l.stdout {
	// 	opts = append(opts, zapcore.AddSync(os.Stdout))
	// }

	if env := strings.ToLower(os.Getenv("RunEnv")); env == "debug" {
		opts = append(opts, zapcore.AddSync(os.Stdout))
	}

	syncWriter := zapcore.NewMultiWriteSyncer(opts...)

	// 自定义时间输出格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("" + t.Format(logTmFmtWithMS) + "")
	}

	// 自定义日志级别显示
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("" + level.CapitalString() + "")
	}

	// 自定义文件：行号输出项
	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// enc.AppendString("[" + l.traceId + "]")
		enc.AppendString("" + caller.TrimmedPath() + "")
	}

	encoderConf := zapcore.EncoderConfig{
		CallerKey:      "caller", // 打印文件名和行数
		LevelKey:       "level",
		MessageKey:     "msg",
		TimeKey:        "time",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,   // 自定义时间格式
		EncodeLevel:    customLevelEncoder,  // 小写编码器
		EncodeCaller:   customCallerEncoder, // 全路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// // level大写染色编码器
	// if l.enableColor {
	// 	encoderConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// }

	// // json 格式化处理
	// if l.jsonFormat {
	// 	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConf),
	// 		syncWriter, zap.NewAtomicLevelAt(l.logMinLevel))
	// }

	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConf),
		syncWriter, zap.NewAtomicLevelAt(zapcore.DebugLevel))
}

func formatField(c context.Context, tag string) []zapcore.Field {
	fields := make([]zapcore.Field, 0)

	if tag != "" {
		fields = append(fields, zap.String("tag", tag))
	}

	if c == nil {
		return fields
	}

	// if g, ok := c.(*gin.Context); ok {
	// 	c = g.Request.Context()
	// }

	var traceID string
	trace := c.Value(ContextKey("x_trace_id"))
	if id, ok := trace.(string); ok {
		traceID = id
	}

	// if traceID == "" {
	// 	traceID = uuid.NewV4().String()
	// }

	return append(fields, zap.String("x_trace_id", traceID))
}

func Ix(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string

	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.Info(msg, fields...)
}

func Ex(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string
	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.Error(msg, fields...)
}

func Dx(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string
	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.Debug(msg, fields...)
}

func Wx(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string
	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.Warn(msg, fields...)
}

func DPx(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string
	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.DPanic(msg, fields...)
}

func Px(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string
	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.Panic(msg, fields...)
}

func Fx(c context.Context, tag string, template interface{}, args ...interface{}) {
	var msg string
	if tpl, flag := template.(string); flag {
		msg = fmt.Sprintf(tpl, args...)
	}

	if tpl, flag := template.(map[string]interface{}); flag {
		msg, _ = jsoniter.MarshalToString(tpl)
	}

	fields := formatField(c, tag)
	logzap.Fatal(msg, fields...)
}
