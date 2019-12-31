/*
 * @Author: calm.wu
 * @Date: 2018-08-17 12:51:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 15:50:28
 */

package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// 使用zap是因为快而不是美

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

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// logFullName: dir/dir/dir/test.log
// maxSize: megabytes, default = 100
// maxAge: 多少天之后变为old file
// maxBackups: old file备份数量
// compress: old file是否压缩tgz
// logLevel: zapcore.DebugLevel
func CreateZapLog(logFullName string, maxSize int, maxAge int, maxBackups int, compress bool,
	logLevel zapcore.Level, callSkip int) *zap.SugaredLogger {

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
		MessageKey:     "M",
		LevelKey:       "L",
		NameKey:        "N",
		TimeKey:        "T",
		CallerKey:      "C",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     timeEncoder, //zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   ShortCallerWithClassFunctionEncoder, //zapcore.ShortCallerEncoder, //
		EncodeName:     zapcore.FullNameEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		w,
		logLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel), zap.AddCallerSkip(callSkip), zap.Development())
	suger := logger.Sugar()
	return suger
}

// InitDefaultZapLog 初始化Zap log
func InitDefaultZapLog(logFullName string, logLevel zapcore.Level, callSkip int) {
	ZLog = CreateZapLog(logFullName, 100, 7, 7, true, logLevel, callSkip)
}

// NewSimpleLog 简化一个logger对象
func NewSimpleLog(out io.Writer) *log.Logger {
	logOutput := out
	if out == nil {
		logOutput = os.Stdout
	}

	return log.New(logOutput, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

// Debug 封装
func Debug(args ...interface{}) {
	if ZLog != nil {
		ZLog.Debug(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Print(composeArgs...)
	}
}

// Debugf 封装
func Debugf(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.Debugf(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Printf(prefix+template, args...)
	}
}

// Info 封装
func Info(args ...interface{}) {
	if ZLog != nil {
		ZLog.Info(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Print(composeArgs...)
	}
}

// Infof 封装
func Infof(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.Infof(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Printf(prefix+template, args...)
	}
}

// Warn 封装
func Warn(args ...interface{}) {
	if ZLog != nil {
		ZLog.Warn(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Print(composeArgs...)
	}
}

// Warnf 封装
func Warnf(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.Warnf(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Printf(prefix+template, args...)
	}
}

// Error 封装
func Error(args ...interface{}) {
	if ZLog != nil {
		ZLog.Error(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Print(composeArgs...)
	}
}

// Errorf 封装
func Errorf(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.Errorf(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Printf(prefix+template, args...)
	}
}

// DPanic 封装
func DPanic(args ...interface{}) {
	if ZLog != nil {
		ZLog.DPanic(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Panic(composeArgs...)
	}
}

// DPanicf 封装
func DPanicf(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.DPanicf(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Panicf(prefix+template, args...)
	}
}

// Panic 封装
func Panic(args ...interface{}) {
	if ZLog != nil {
		ZLog.Panic(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Panic(composeArgs...)
	}
}

// Panicf 封装
func Panicf(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.Panicf(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Panicf(prefix+template, args...)
	}
}

// Fatal 封装
func Fatal(args ...interface{}) {
	if ZLog != nil {
		ZLog.Fatal(args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		composeArgs := []interface{}{prefix}
		composeArgs = append(composeArgs, args...)
		log.Fatal(composeArgs...)
	}
}

// Fatalf 封装
func Fatalf(template string, args ...interface{}) {
	if ZLog != nil {
		ZLog.Fatalf(template, args...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		prefix := fmt.Sprintf("%v:%v: ", path.Base(file), line)
		log.Fatalf(prefix+template, args...)
	}
}
