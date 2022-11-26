package logzap

import (
	"context"
	"fmt"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logZap *zap.Logger

type ContextKey string

type Logger struct {
	Filename   string // ⽇志⽂件路径
	MaxSize    int    // 单位为MB,默认为100MB
	MaxAge     int    // 文件最多保存多少天
	MaxBackups int
}

var logConf *Logger

func Zap() *zap.Logger {
	return logZap
}

func InitZap(name string) {
	if logConf == nil {
		logConf = &Logger{}
	}

	if logConf.Filename == "" {
		logPath := "/log/"
		env := strings.ToLower(os.Getenv("envType"))
		if env != "product" && env != "test" && env != "gray" {
			logPath = "runtime/"
		}

		logPath = fmt.Sprintf("%s/%s.log", logPath, name)

		if closed := viper.GetBool("log.closed"); closed {
			logPath = "/dev/null"
		}

		logConf.Filename = logPath
	}

	if logConf.MaxSize == 0 {
		maxSize := viper.GetInt("log.max_size")
		if maxSize == 0 {
			maxSize = 5120
		}

		logConf.MaxSize = maxSize
	}

	if logConf.MaxAge == 0 {
		maxAge := viper.GetInt("log.max_age")
		if maxAge == 0 {
			maxAge = 7
		}
		logConf.MaxAge = maxAge
	}

	if logConf.MaxBackups == 0 {
		maxBackup := viper.GetInt("log.max_backup")
		if maxBackup == 0 { // 如果未设置则默认保留5个
			maxBackup = 5
		}

		if maxBackup < 0 { // 如果设置为 -1 代表保留全部
			maxBackup = 0
		}

		logConf.MaxBackups = maxBackup
	}

	logZap = zap.New(initCore(logConf), zap.AddCallerSkip(1), zap.AddCaller())
}

func initCore(config *Logger) zapcore.Core {
	opts := []zapcore.WriteSyncer{
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.Filename, // ⽇志⽂件路径
			MaxSize:    config.MaxSize,  // 单位为MB,默认为100MB
			MaxAge:     config.MaxAge,   // 文件最多保存多少天
			LocalTime:  true,            // 采用本地时间
			Compress:   false,           // 是否压缩日志
			MaxBackups: config.MaxBackups,
		}),
	}

	// if l.stdout {
	// 	opts = append(opts, zapcore.AddSync(os.Stdout))
	// }

	syncWriter := zapcore.NewMultiWriteSyncer(opts...)

	// 自定义时间输出格式
	// customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 	enc.AppendString(t.Format(logTmFmtWithMS))
	// 	// enc.AppendString(strconv.Itoa(int(t.UnixNano() / 1e6)))
	// }

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
		CallerKey:      "caller_line", // 打印文件名和行数
		LevelKey:       "level",
		MessageKey:     "msg",
		TimeKey:        "@timestamp",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.RFC3339TimeEncoder, // 自定义时间格式
		EncodeLevel:    customLevelEncoder,         // 小写编码器
		EncodeCaller:   customCallerEncoder,        // 全路径编码器
		EncodeDuration: zapcore.MillisDurationEncoder,
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

	requestURI := c.Value(ContextKey("request_uri"))
	if uri, ok := requestURI.(string); ok && uri != "" {
		fields = append(fields, zap.String("request_uri", uri))
	}

	// var traceID string
	trace := c.Value(ContextKey("x_trace_id"))
	if traceID, ok := trace.(string); ok {
		fields = append(fields, zap.String("x_trace_id", traceID))
	}

	return fields
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
	logZap.Info(msg, fields...)
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
	logZap.Error(msg, fields...)
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
	logZap.Debug(msg, fields...)
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
	logZap.Warn(msg, fields...)
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
	logZap.DPanic(msg, fields...)
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
	logZap.Panic(msg, fields...)
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
	logZap.Fatal(msg, fields...)
}
