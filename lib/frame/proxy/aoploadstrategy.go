package proxy

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
)

type AopLoadStrategy struct {
}

func (h *AopLoadStrategy) LoadInstance(local *context.LocalStack, target ProxyTarger,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target == nil {
		return false
	}
	if f, ok := target.(ProxyFilter); ok {
		AddAopProxyFilter(f)
		return true
	}
	return false

}

func (h *AopLoadStrategy) Order() int {
	return 50
}

func (h *AopLoadStrategy) ProxyTarget() *ProxyClass {
	return nil
}

var aopLoadStrategy AopLoadStrategy = AopLoadStrategy{}

func GetAopLoadStrategy() *AopLoadStrategy {
	return &aopLoadStrategy
}

func init() {
	application.AddProxyInstance("", ProxyTarger(&aopLoadStrategy))
}
