package logzap

import (
	"context"
	"os"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

var sugar *zap.SugaredLogger

// Init 11
func init() {
	// until := time.Now().Add(5 * time.Second)
	// AppConfig := config.GetConfig()
	// for AppConfig == nil {
	// 	if time.Now().After(until) {
	// 		break
	// 	}

	// 	fmt.Println("config not init, sleep...")
	// 	time.Sleep(time.Second)
	// }

	errlog := viper.GetString("log.error")

	isDebug := false
	env := viper.GetString("env")
	if env == "debug" {
		isDebug = true
	}

	// fmt.Println("error log path:", errlog)
	initLog(errlog, isDebug)
}

// InitLog 初始化
func initLog(path string, isDebug bool) {
	core := initCore(path, true)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()

	defer logger.Sync()
}

func initCore(path string, isDebug bool) zapcore.Core {
	opts := []zapcore.WriteSyncer{
		zapcore.AddSync(&lumberjack.Logger{
			Filename:  path,  // ⽇志⽂件路径
			MaxSize:   512,   // 单位为MB,默认为512MB
			MaxAge:    7,     // 文件最多保存多少天
			LocalTime: true,  // 采用本地时间
			Compress:  false, // 是否压缩日志
		}),
	}

	// if l.stdout {
	// 	opts = append(opts, zapcore.AddSync(os.Stdout))
	// }

	if isDebug {
		opts = append(opts, zapcore.AddSync(os.Stdout))
	}

	syncWriter := zapcore.NewMultiWriteSyncer(opts...)

	// 自定义时间输出格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}

	// 自定义日志级别显示
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + level.CapitalString() + "]")
	}

	// 自定义文件：行号输出项
	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// enc.AppendString("[" + l.traceId + "]")
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	encoderConf := zapcore.EncoderConfig{
		CallerKey:      "caller_line", // 打印文件名和行数
		LevelKey:       "level_name",
		MessageKey:     "msg",
		TimeKey:        "ts",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,   // 自定义时间格式
		EncodeLevel:    customLevelEncoder,  // 大小写编码器
		EncodeCaller:   customCallerEncoder, // 全路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// // level大写染色编码器
	// if l.enableColor {
	// encoderConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// }

	if isDebug {
		encoderConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// // json 格式化处理
	// if l.jsonFormat {
	// 	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConf),
	// 		syncWriter, zap.NewAtomicLevelAt(l.logMinLevel))
	// }

	// return zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConf),
	// syncWriter, zap.NewAtomicLevelAt(zapcore.DebugLevel))
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
	trace := c.Value("x_trace_id")
	if id, ok := trace.(string); ok {
		traceID = id
	}

	if traceID == "" {
		traceID = uuid.NewV4().String()
	}

	return append(fields, zap.String("x_trace_id", traceID))
}

// I info
func I(template string, args ...interface{}) {
	// if sugar == nil {
	// 	initLog()
	// }

	sugar.Infof(template, args...)
}

// E Error
func E(template string, args ...interface{}) {
	// if sugar == nil {
	// 	InitLog()
	// }

	sugar.Errorf(template, args...)
}

func Dx(c context.Context, tag, template string, args ...interface{}) {
	// msg := fmt.Sprintf(template, args...)
	// fields := formatField(c, tag)

	// logzap.Debug(msg, fields...)
	sugar.Debugf(template, args...)
}

func Wx(c context.Context, tag, template string, args ...interface{}) {
	// msg := fmt.Sprintf(template, args...)
	// fields := formatField(c, tag)

	// logzap.Warn(msg, fields...)
	sugar.Warnf(template, args...)
}

func DPx(c context.Context, tag, template string, args ...interface{}) {
	// msg := fmt.Sprintf(template, args...)
	// fields := formatField(c, tag)

	// logzap.DPanic(msg, fields...)
	sugar.DPanicf(template, args...)
}

func Px(c context.Context, tag, template string, args ...interface{}) {
	// msg := fmt.Sprintf(template, args...)
	// fields := formatField(c, tag)

	// logzap.Panic(msg, fields...)
	sugar.Panicf(template, args...)
}

func Fx(c context.Context, tag, template string, args ...interface{}) {
	// msg := fmt.Sprintf(template, args...)
	// fields := formatField(c, tag)

	// logzap.Fatal(msg, fields...)
	sugar.Fatalf(template, args...)
}
