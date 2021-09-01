package http

import (
	"context"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

type HttpControllerLoadStrategy struct {
	logger application.AppLoger `FrameAutowired:""`
}

func (h *HttpControllerLoadStrategy) LoadInstance(local context.Context, target *application.DynamicProxyInstanceNode,
	application *application.Application,
	applicationContext *application.ApplicationContext) bool {

	if target == nil {
		return false
	}
	target := target1.Target
	if target == nil || target.ProxyTarget() == nil {
		return false
	}
	if len(target.ProxyTarget().Annotations) == 0 {
		return false
	}
	httpAnno := GetRequestAnnotation(target.ProxyTarget().Annotations)
	if httpAnno == nil {
		return false
	}
	GetFrameHttpFactory().AddControllerProxyTarget(local, target, applicationContext)
	return true
}

func (h *HttpControllerLoadStrategy) Order() int {
	return 100
}

func (h *HttpControllerLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var httpControllerLoadStrategy HttpControllerLoadStrategy = HttpControllerLoadStrategy{}

type HttpFilterLoadStrategy struct {
}

func (h *HttpFilterLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
		return false
	}
	target := target1.Target
	if f, ok := target.(Filter); ok {
		AddFilter(f)
		return true
	}
	return false
}

func (h *HttpFilterLoadStrategy) Order() int {
	return 90
}

func (h *HttpFilterLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var httpFilterLoadStrategy HttpFilterLoadStrategy = HttpFilterLoadStrategy{}

type HttpComponentLoadStrategy struct {
}

func (h *HttpComponentLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
		return false
	}
	target := target1.Target
	if f, ok := target.(HttpViewRender); ok {
		AddHttpViewRender(f)
		return true
	}
	return false
}

func (h *HttpComponentLoadStrategy) Order() int {
	return 80
}

func (h *HttpComponentLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var httpComponentLoadStrategy HttpComponentLoadStrategy = HttpComponentLoadStrategy{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&httpControllerLoadStrategy))
	application.AddProxyInstance("", proxyclass.ProxyTarger(&httpFilterLoadStrategy))
	application.AddProxyInstance("", proxyclass.ProxyTarger(&httpComponentLoadStrategy))
}
