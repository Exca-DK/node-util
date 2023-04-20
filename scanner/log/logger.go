package log

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogField struct {
	key, value string
}

func NewTField[T any](key string, val T) LogField {
	return LogField{key: key, value: fmt.Sprintf("%v", val)}
}

func NewStringField(key, val string) LogField {
	return LogField{key: key, value: val}
}

func NewByteField(key string, val byte) LogField {
	return LogField{key: key, value: fmt.Sprintf("%v", val)}
}

func NewErrorField(err error) LogField {
	var msg string
	if err == nil {
		msg = "nil"
	} else {
		msg = err.Error()
	}
	return LogField{key: "error", value: msg}
}

type Logger interface {
	Trace(string, ...LogField)
	Debug(string, ...LogField)
	Info(string, ...LogField)
	Warn(string, ...LogField)
	Error(string, ...LogField)
	Fatal(string, ...LogField)
}

var currLvl zapcore.Level = zap.InfoLevel
var root atomic.Value

type LoggingLvl int8

const (
	TRACE LoggingLvl = LoggingLvl(zapcore.DebugLevel - 1)
	DEBUG LoggingLvl = TRACE + 1
	INFO  LoggingLvl = DEBUG + 1
	WARN  LoggingLvl = INFO + 1
	ERROR LoggingLvl = WARN + 1
	FATAL LoggingLvl = LoggingLvl(zapcore.FatalLevel)
)

func SetLoggingLvl(lvl LoggingLvl) {
	currLvl = zapcore.Level(lvl)
}

func Root() Logger {
	if val := root.Load(); val != nil {
		return val.(Logger)
	}
	root.CompareAndSwap(nil, NewLogger())
	return root.Load().(Logger)
}

func NewLogger() Logger {
	return &logger{inner: newDevZapLogger(zap.NewAtomicLevelAt(currLvl))}
}

func NewLoggerWithId(id string) Logger {
	return &logger{id: id, inner: newDevZapLogger(zap.NewAtomicLevelAt(currLvl))}
}

func newDevZapLogger(lvl zapcore.LevelEnabler) *zap.Logger {
	encCfg := zap.NewDevelopmentEncoderConfig()
	encCfg.EncodeLevel = capitalColorLevelEncoder
	zapCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encCfg), zapcore.AddSync(os.Stderr), lvl)
	return zap.New(zapCore, zap.AddCallerSkip(2))
}

func capitalColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if level == zapcore.Level(TRACE) {
		enc.AppendString(color.CyanString("TRACE"))
		return
	}
	zapcore.CapitalLevelEncoder(level, enc)
}

type logger struct {
	id    string
	inner *zap.Logger
}

func (l *logger) Trace(msg string, fields ...LogField) {
	l.log(zapcore.Level(TRACE), msg, fields...)
}

func (l *logger) Debug(msg string, fields ...LogField) {
	l.log(zap.DebugLevel, msg, fields...)
}

func (l *logger) Info(msg string, fields ...LogField) {
	l.log(zap.InfoLevel, msg, fields...)
}

func (l *logger) Warn(msg string, fields ...LogField) {
	l.log(zap.WarnLevel, msg, fields...)
}

func (l *logger) Error(msg string, fields ...LogField) {
	l.log(zap.ErrorLevel, msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...LogField) {
	l.log(zap.FatalLevel, msg, fields...)
}

func (l *logger) log(lvl zapcore.Level, msg string, _fields ...LogField) {
	fields := make([]zapcore.Field, len(_fields))
	for i, field := range _fields {
		fields[i] = zap.String(field.key, field.value)
	}
	if len(l.id) > 0 {
		l.inner.Log(lvl, fmt.Sprintf("id:%v-msg:%v", l.id, msg), fields...)
	} else {
		l.inner.Log(lvl, msg, fields...)
	}
}
