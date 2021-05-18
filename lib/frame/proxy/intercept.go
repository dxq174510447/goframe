package proxy

import (
	"fmt"
	"goframe/lib/frame/context"
	"reflect"
)

// ProxyFilterFactory
// 每次创建filter都要创建两个类A，B, A用来实现ProxyFilter接口，B用来实现ProxyFilterFactory接口创建A
type ProxyFilterFactory interface {
	GetInstance(map[string]interface{}) ProxyFilter

	AnnotationMatch() []string
}

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
	return 999999
}

var defaultProxyFilter DefaultProxyFilter = DefaultProxyFilter{}

type DefaultProxyFilterFactory struct {
}

func (d *DefaultProxyFilterFactory) GetInstance(m map[string]interface{}) ProxyFilter {
	r1 := ProxyFilter(&defaultProxyFilter)
	return r1
}

func (d *DefaultProxyFilterFactory) AnnotationMatch() []string {
	return nil
}

var defaultProxyFilterFactory ProxyFilterFactory = ProxyFilterFactory(&DefaultProxyFilterFactory{})

func init() {
	AddDefaultInvokerFilterFactory(defaultProxyFilterFactory)
}
