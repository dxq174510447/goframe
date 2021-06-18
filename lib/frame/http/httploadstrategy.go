package http

import (
	"encoding/json"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

type HttpControllerLoadStrategy struct {
	Logger    application.AppLoger `FrameAutowired:""`
	SerConfig *ServerConfig        `FrameValue:"${server}"`
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
	if DefaultServConfig == nil {
		if h.Logger.IsDebugEnable() {
			s, _ := json.Marshal(h.SerConfig)
			h.Logger.Debug(local, "httpConfig %s", string(s))
		}
		DefaultServConfig = h.SerConfig
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
