package strategy

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"reflect"
)

// DefaultProxyFilter 默认可执行拦截器
type DefaultProxyFilter struct {
}

func (d *DefaultProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {
	return (*invoker).Call(arg)
}

func (d *DefaultProxyFilter) Order() int {
	return 99999999
}

func (d *DefaultProxyFilter) AnnotationMatch() []string {
	return nil
}

func (d *DefaultProxyFilter) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var defaultProxyFilter DefaultProxyFilter = DefaultProxyFilter{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&defaultProxyFilter))
}
