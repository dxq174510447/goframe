package event

import (
	"context"
	"github.com/dxq174510447/goframe/lib/frame/application"
)

type FrameListenerLoadStrategy struct {
	logger application.AppLoger `FrameAutowired:""`
}

func (h *FrameListenerLoadStrategy) LoadInstance(local context.Context,
	target *application.DynamicProxyInstanceNode,
	application *application.Application,
	applicationContext *application.ApplicationContext) bool {

	if target == nil {
		return false
	}

	if f, ok := target.Target.(FrameListener); ok {
		GetFrameEventDispatcher().AddEventListener(local, f)
		return false
	}

	return false
}

func (h *FrameListenerLoadStrategy) Order() int {
	return 100
}

var frameListenerLoadStrategy FrameListenerLoadStrategy = FrameListenerLoadStrategy{}

func init() {
	application.GetResourcePool().RegisterInstance("", &frameListenerLoadStrategy)
}
