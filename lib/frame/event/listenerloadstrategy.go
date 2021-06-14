package event

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

type FrameListenerLoadStrategy struct {
}

func (h *FrameListenerLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
		return false
	}
	target := target1.Target
	if f, ok := target.(FrameListener); ok {
		core.AddClassProxy(target)
		GetFrameEventDispatcher().AddEventListener(local, f)
		return true
	}
	return false

}

func (h *FrameListenerLoadStrategy) Order() int {
	return 100
}

func (h *FrameListenerLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var frameListenerLoadStrategy FrameListenerLoadStrategy = FrameListenerLoadStrategy{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&frameListenerLoadStrategy))
}
