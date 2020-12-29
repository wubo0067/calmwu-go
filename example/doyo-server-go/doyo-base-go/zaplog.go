/*
 * @Author: calm.wu
 * @Date: 2018-08-17 12:51:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 15:50:28
 */

package base

import (
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	ZLog *zap.SugaredLogger
)

func ShortCallerWithClassFunctionEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	path := caller.TrimmedPath()
	if f := runtime.FuncForPC(caller.PC); f != nil {
		name := f.Name()
		i := strings.LastIndex(name, "/")
		j := strings.Index(name[i+1:], ".")
		path += " " + name[i+j+2:]
	}
	enc.AppendString(path)
}

// logFullName: dir/dir/dir/test.log
// maxSize: megabytes, default = 100
// maxAge: 多少天之后变为old file
// maxBackups: old file备份数量
// compress: old file是否压缩tgz
// logLevel: zapcore.DebugLevel
func CreateZapLog(logFullName string, maxSize int, maxAge int, maxBackups int, compress bool,
	logLevel zapcore.Level) *zap.SugaredLogger {

	if maxSize < 100 {
		maxSize = 100
	}

	if maxAge < 0 {
		maxAge = 0
	}

	if maxBackups < 0 {
		maxBackups = 0
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFullName,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // days
		Compress:   compress,
	})

	cfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "name",
		TimeKey:        "ts",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   ShortCallerWithClassFunctionEncoder, //zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		w,
		logLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	suger := logger.Sugar()
	return suger
}

func InitDefaultZapLog(logFullName string, logLevel zapcore.Level) {
	ZLog = CreateZapLog(logFullName, 100, 7, 7, true, logLevel)
}

func NewSimpleLog(out io.Writer) *log.Logger {
	logOutput := out
	if out == nil {
		logOutput = os.Stderr
	}

	return log.New(logOutput, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}
