package strategy

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

type AopLoadStrategy struct {
}

func (h *AopLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
		return false
	}
	target := target1.Target
	if f, ok := target.(proxyclass.ProxyFilter); ok {
		core.AddAopProxyFilter(f)
		return true
	}
	return false

}

func (h *AopLoadStrategy) Order() int {
	return 50
}

func (h *AopLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var aopLoadStrategy AopLoadStrategy = AopLoadStrategy{}

func GetAopLoadStrategy() *AopLoadStrategy {
	return &aopLoadStrategy
}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&aopLoadStrategy))
}
