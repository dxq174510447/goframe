package event

import (
	"context"
	"goframe/lib/frame/application"
)

type FrameListenerLoadStrategy struct {
	logger application.AppLoger `FrameAutowired:""`
}

func (h *FrameListenerLoadStrategy) LoadInstance(local context.Context,
	target *application.DynamicProxyInstanceNode,
	application *application.Application,
	applicationContext *application.ApplicationContext) (bool, error) {

	if target == nil {
		return false, nil
	}

	if f, ok := target.Target.(FrameListener); ok {
		GetFrameEventDispatcher().AddEventListener(local, f)
		return true, nil
	}

	return false, nil
}

func (h *FrameListenerLoadStrategy) Order() int {
	return 100
}

func init() {
	frameListenerLoadStrategy := FrameListenerLoadStrategy{}
	_ = application.LoadInstanceHandler(&frameListenerLoadStrategy)

	application.GetResourcePool().RegisterInstance("", &frameListenerLoadStrategy)
}
