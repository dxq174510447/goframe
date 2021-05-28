package proxy

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"reflect"
)

// ProxyFilter 代理接口
type ProxyFilter interface {

	// Execute
	// context 上下文
	// classInfo 代理对象类
	// methodInfo 代理方法属性
	// invoker 可执行接口
	// arg 参数
	Execute(context *context.LocalStack,
		classInfo *ProxyClass,
		methodInfo *ProxyMethod,
		invoker *reflect.Value, arg []reflect.Value) []reflect.Value

	// SetNext 下一个
	SetNext(next ProxyFilter)

	// Order 顺序
	Order() int

	AnnotationMatch() []string
}

// DefaultProxyFilter 默认可执行拦截器
type DefaultProxyFilter struct {
}

func (d *DefaultProxyFilter) Execute(stack *context.LocalStack,
	classInfo *ProxyClass,
	methodInfo *ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value) []reflect.Value {
	fmt.Println("default begin")
	defer fmt.Println("default end")
	return (*invoker).Call(arg)
}

func (d *DefaultProxyFilter) SetNext(next ProxyFilter) {

}

func (d *DefaultProxyFilter) Order() int {
	return 99999999
}

func (d *DefaultProxyFilter) AnnotationMatch() []string {
	return nil
}

func (d *DefaultProxyFilter) ProxyTarget() *ProxyClass {
	return nil
}

var defaultProxyFilter DefaultProxyFilter = DefaultProxyFilter{}

func init() {
	application.AddProxyInstance("", ProxyTarger(&defaultProxyFilter))
}
