package http

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/ctx"
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

	RestControllerAnnotationName = "RestController"

	RequestMappingAnnotationName = "RequestMapping"

	RequestParamAnnotationName = "RequestParam"

	RequestBodyAnnotationName = "RequestBody"

	CookieValueAnnotationName = "CookieValue"

	PathVariableAnnotationName = "PathVariable"

	RequestHeaderAnnotationName = "RequestHeader"
)

func SetCurrentControllerInvoker(local context.Context, invoker1 *ControllerVar) {
	ctx.WithValue(local, CurrentControllerInvoker, invoker1)
}
func GetCurrentControllerInvoker(local context.Context) *ControllerVar {
	invoker := local.Value(CurrentControllerInvoker)
	return invoker.(*ControllerVar)
}

func SetCurrentHttpRequest(local context.Context, request *http.Request) {
	ctx.WithValue(local, CurrentHttpRequest, request)
}
func GetCurrentHttpRequest(local context.Context) *http.Request {
	invoker := local.Value(CurrentHttpRequest)
	return invoker.(*http.Request)
}

func SetCurrentHttpResponse(local context.Context, response http.ResponseWriter) {
	ctx.WithValue(local, CurrentHttpResponse, response)
}
func GetCurrentHttpResponse(local context.Context) http.ResponseWriter {
	invoker := local.Value(CurrentHttpResponse)
	return invoker.(http.ResponseWriter)
}

func SetCurrentFilterIndex(local context.Context, index int) {
	ctx.WithValue(local, FilterIndexWaitToExecute, index)
}

func GetCurrentFilterIndex(local context.Context) int {
	index := local.Value(FilterIndexWaitToExecute)
	if index == nil {
		return 0
	}
	return index.(int)
}

func GetRequestAnnotationSetting(annotations []*proxyclass.AnnotationClass) *RestAnnotationSetting {
	for _, annotation := range annotations {
		if annotation.Name == AnnotationRestController || annotation.Name == AnnotationController {
			if r, ok := annotation.Value[AnnotationValueRestKey]; ok {
				return r.(*RestAnnotationSetting)
			}
			return nil
		}
	}
	return nil
}
func GetRequestAnnotation(annotations []*proxyclass.AnnotationClass) *proxyclass.AnnotationClass {
	for _, annotation := range annotations {
		if annotation.Name == AnnotationRestController || annotation.Name == AnnotationController {
			return annotation
		}
	}
	return nil
}

func NewRestAnnotation(httpPath string,
	httpMethod string,
	methodParameter string,
	pathVariable string,
	headerParameter string,
	methodRender string) *proxyclass.AnnotationClass {
	return &proxyclass.AnnotationClass{
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
