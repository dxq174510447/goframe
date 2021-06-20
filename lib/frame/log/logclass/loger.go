package logclass

import (
	"github.com/dxq174510447/goframe/lib/frame/context"
	"reflect"
	"text/template"
)

// 和日志相关到接口

type AppLogFactoryer interface {
	Parse(content string, funcMap template.FuncMap)

	// GetLoggerType p 非指针类型
	GetLoggerType(p reflect.Type) AppLoger

	GetLoggerString(className string) AppLoger
}

type AppLoger interface {
	Trace(local *context.LocalStack, format string, a ...interface{})

	IsTraceEnable() bool

	Debug(local *context.LocalStack, format string, a ...interface{})

	IsDebugEnable() bool

	Info(local *context.LocalStack, format string, a ...interface{})

	IsInfoEnable() bool

	Warn(local *context.LocalStack, format string, a ...interface{})

	IsWarnEnable() bool

	//err 可空
	Error(local *context.LocalStack, err interface{}, format string, a ...interface{})

	IsErrorEnable() bool
}
