package http

import (
	"encoding/json"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/event"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"net/http"
	"time"
)

type ServerServletConfig struct {
	ContextPath string
}
type ServerConfig struct {
	Port    int
	Servlet *ServerServletConfig
}

type HttpServListener struct {
	Logger     application.AppLoger        `FrameAutowired:""`
	Dispatcher *event.FrameEventDispatcher `FrameAutowired:""`
}

func (h *HttpServListener) Starting(local *context.LocalStack) {

}

func (h *HttpServListener) EnvironmentPrepared(local *context.LocalStack, environment *application.ConfigurableEnvironment) {

}

func (h *HttpServListener) Running(local *context.LocalStack, application *application.FrameApplicationContext) {
	var setting *ServerConfig = &ServerConfig{}
	application.Environment.GetObjectValue("server", setting)

	if h.Logger.IsDebugEnable() {
		s, _ := json.Marshal(setting)
		h.Logger.Debug(local, "httpConfig %s", string(s))
	}

	DefaultServConfig = setting

	var address string = fmt.Sprintf("%s:%d", "", setting.Port)
	h.Logger.Info(local, "http开始监听 %s", address)

	go func() {
		l := context.NewLocalStack()
		l.SetThread()

		select {
		case <-time.After(5 * time.Second):
			h.Logger.Debug(l, "http事件初始化")
		}

		e := &WebServletStartedEvent{}
		h.Dispatcher.DispatchEvent(local, e)
	}()

	http.ListenAndServe(address, nil)

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
}
