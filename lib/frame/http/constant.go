package http

import (
	"bytes"
	"firstgo/frame/context"
	"firstgo/frame/proxy"
	"fmt"
	"net/http"
	"runtime"
)

const (
	AnnotationRestController = "AnnotationRestController_"

	AnnotationController = "AnnotationController_"

	AnnotationValueRestKey = "AnnotationValueRestKey_"

	FilterIndexWaitToExecute = "FilterIndexWaitToExecute_"

	CurrentControllerInvoker = "CurrentControllerInvoker_"

	CurrentHttpRequest = "CurrentHttpRequest_"

	CurrentHttpResponse = "CurrentHttpResponse_"
)

func SetCurrentControllerInvoker(local *context.LocalStack, invoker1 *ControllerVar) {
	local.Set(CurrentControllerInvoker, invoker1)
}
func GetCurrentControllerInvoker(local *context.LocalStack) *ControllerVar {
	invoker := local.Get(CurrentControllerInvoker)
	return invoker.(*ControllerVar)
}

func SetCurrentHttpRequest(local *context.LocalStack, request *http.Request) {
	local.Set(CurrentHttpRequest, request)
}
func GetCurrentHttpRequest(local *context.LocalStack) *http.Request {
	invoker := local.Get(CurrentHttpRequest)
	return invoker.(*http.Request)
}

func SetCurrentHttpResponse(local *context.LocalStack, response http.ResponseWriter) {
	local.Set(CurrentHttpResponse, response)
}
func GetCurrentHttpResponse(local *context.LocalStack) http.ResponseWriter {
	invoker := local.Get(CurrentHttpResponse)
	return invoker.(http.ResponseWriter)
}

func SetCurrentFilterIndex(local *context.LocalStack, index int) {
	local.Set(FilterIndexWaitToExecute, index)
}

func GetCurrentFilterIndex(local *context.LocalStack) int {
	index := local.Get(FilterIndexWaitToExecute)
	if index == nil {
		return 0
	}
	return index.(int)
}

func GetRequestAnnotationSetting(annotations []*proxy.AnnotationClass) *RestAnnotationSetting {
	for _, annotation := range annotations {
		if annotation.Name == AnnotationRestController {
			r, _ := annotation.Value[AnnotationValueRestKey]
			return r.(*RestAnnotationSetting)
		}
	}
	return nil
}

func NewRestAnnotation(httpPath string,
	httpMethod string,
	methodParameter string,
	pathVariable string,
	headerParameter string,
	methodRender string) *proxy.AnnotationClass {
	return &proxy.AnnotationClass{
		Name: AnnotationRestController,
		Value: map[string]interface{}{
			AnnotationValueRestKey: &RestAnnotationSetting{
				HttpPath:        httpPath,
				HttpMethod:      httpMethod,
				QueryParameter:  methodParameter,
				HeaderParameter: headerParameter,
				PathVariable:    pathVariable,
				MethodRender:    methodRender,
			},
		},
	}
}

func PrintStackTrace(err interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", err)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}

func GetControllerPathPrefix(dispatchServlet *DispatchServlet, target proxy.ProxyTarger) string {
	//context-path
	var sp string = dispatchServlet.ContextPath
	if sp == "/" {
		sp = ""
	}

	//controller-path
	var classRestSetting *RestAnnotationSetting = GetRequestAnnotationSetting(target.ProxyTarget().Annotations)
	var cp string = classRestSetting.HttpPath
	if cp == "/" {
		cp = ""
	}
	return fmt.Sprintf("%s%s", sp, cp)
}
