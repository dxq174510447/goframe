package proxyclass

import (
	"github.com/dxq174510447/goframe/lib/frame/context"
	"reflect"
)

type AnnotationClass struct {
	Name  string
	Value map[string]interface{}
}

type ProxyClass struct {
	Name        string
	Target      interface{}
	Methods     []*ProxyMethod
	Annotations []*AnnotationClass
}

type ProxyMethod struct {
	Name        string
	Annotations []*AnnotationClass
}

// ProxyTarger 代理对象都必须继承实现的方法
type ProxyTarger interface {
	ProxyTarget() *ProxyClass
}

type ProxyFilterWrapper struct {
	Target ProxyFilter
	Next   *ProxyFilterWrapper
}

func (p *ProxyFilterWrapper) Execute(context *context.LocalStack,
	classInfo *ProxyClass,
	methodInfo *ProxyMethod,
	invoker *reflect.Value, arg []reflect.Value) []reflect.Value {
	return p.Target.Execute(context, classInfo, methodInfo, invoker, arg, p.Next)
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
		invoker *reflect.Value, arg []reflect.Value, next *ProxyFilterWrapper) []reflect.Value

	// Order 顺序
	Order() int

	AnnotationMatch() []string
}

func NewSingleAnnotation(annotationName string, value map[string]interface{}) *AnnotationClass {
	return &AnnotationClass{
		Name:  annotationName,
		Value: value,
	}
}
