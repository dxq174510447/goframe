package http

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

type HttpControllerLoadStrategy struct {
}

func (h *HttpControllerLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
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
	AddControllerProxyTarget(target)
	return true
}

func (h *HttpControllerLoadStrategy) Order() int {
	return 100
}

func (h *HttpControllerLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var httpControllerLoadStrategy HttpControllerLoadStrategy = HttpControllerLoadStrategy{}

func GetHttpControllerLoadStrategy() *HttpControllerLoadStrategy {
	return &httpControllerLoadStrategy
}

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

func GetHttpFilterLoadStrategy() *HttpFilterLoadStrategy {
	return &httpFilterLoadStrategy
}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&httpControllerLoadStrategy))
	application.AddProxyInstance("", proxyclass.ProxyTarger(&httpFilterLoadStrategy))
}
