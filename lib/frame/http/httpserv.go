package http

import (
	"encoding/json"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"net/http"
)

type ServerServletConfig struct {
	ContextPath string
}
type ServerConfig struct {
	Port    int
	Servlet *ServerServletConfig
}

type HttpServListener struct {
	Logger application.AppLoger `FrameAutowired:""`
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
