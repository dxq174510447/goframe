package http

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/event"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"net/http"
	"time"
)

type ServerServletConfig struct {
	ContextPath string
}

type TestHttpStart struct {
	Logger logclass.AppLoger `FrameAutowired:""`
}

func (t *TestHttpStart) HttpStart(local *context.LocalStack, applicationContext *application.FrameApplicationContext) {
	t.Logger.Debug(local, "测试类启动 不启动http请求  模拟http请求启动")
}

func (t *TestHttpStart) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var testHttpStart TestHttpStart = TestHttpStart{}

func GetTestHttpStart() *TestHttpStart {
	return &testHttpStart
}

type HttpServListener struct {
	Logger     logclass.AppLoger           `FrameAutowired:""`
	Dispatcher *event.FrameEventDispatcher `FrameAutowired:""`
	SerConfig  *ServerConfig               `FrameValue:"${server}"`
}

func (h *HttpServListener) Starting(local *context.LocalStack) {

}

func (h *HttpServListener) EnvironmentPrepared(local *context.LocalStack, environment *application.ConfigurableEnvironment) {

}

func (h *HttpServListener) Running(local *context.LocalStack, application *application.FrameApplicationContext) {

	go func() {
		l := context.NewLocalStack()
		l.SetThread()

		select {
		case <-time.After(4 * time.Second):
			h.Logger.Debug(l, "http事件初始化")
		}

		e := &WebServletStartedEvent{}
		h.Dispatcher.DispatchEvent(local, e)
	}()

	if application.CustomerStarter != nil {
		h.Logger.Info(local, "http使用自带http初始化")
		application.CustomerStarter.HttpStart(local, application)
	} else {

		for _, r := range GetDispatchServlet().GetRouteMapping() {
			h.Logger.Debug(local, "http路径和处理控制器 %s  %s", r.Path, r.Invoker.Target.ProxyTarget().Name)
			http.HandleFunc(r.Path, r.Handler)
		}

		var address string = fmt.Sprintf("%s:%d", "", h.SerConfig.Port)
		h.Logger.Info(local, "http开始监听 %s", address)
		http.ListenAndServe(address, nil)
	}

}

func (h *HttpServListener) Failed(local *context.LocalStack, application *application.FrameApplicationContext, err interface{}) {

}

func (h *HttpServListener) Order() int {
	return 999999
}

func (h *HttpServListener) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var httpServListener HttpServListener = HttpServListener{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&httpServListener))
	application.AddProxyInstance("", proxyclass.ProxyTarger(&testHttpStart))
}
