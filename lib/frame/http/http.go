package http

import (
	"encoding/json"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/exception"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"github.com/dxq174510447/goframe/lib/frame/vo"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type RouteMapping struct {
	Path         string
	Handler      func(http.ResponseWriter, *http.Request)
	AbsolutePath bool
	Invoker      *ControllerVar
}
type ServerConfig struct {
	Port    int
	Servlet *ServerServletConfig
}

type HttpViewRender interface {
	SuccessRender(local *context.LocalStack, controller *ControllerVar,
		proxyMethod *proxyclass.ProxyMethod, methodRequestSetting *RestAnnotationSetting,
		request *http.Request, response http.ResponseWriter, result interface{})
	ErrorRender(local *context.LocalStack, controller *ControllerVar,
		proxyMethod *proxyclass.ProxyMethod, methodRequestSetting *RestAnnotationSetting,
		request *http.Request, response http.ResponseWriter, throwable interface{})
}

type DefaultHttpViewRender struct {
}

func (d *DefaultHttpViewRender) SuccessRender(local *context.LocalStack, controller *ControllerVar,
	proxyMethod *proxyclass.ProxyMethod, methodRequestSetting *RestAnnotationSetting,
	request *http.Request, response http.ResponseWriter, result interface{}) {

	response.Header().Add("Content-Type", "application/json;charset=UTF-8")

	var successJson *vo.JsonResult
	if result != nil {
		if reflect.TypeOf(result).Kind() == reflect.Slice {
			successJson = util.JsonUtil.BuildJsonArraySuccess(result, -1)
		} else {
			successJson = util.JsonUtil.BuildJsonSuccess(result)
		}
	} else {
		successJson = util.JsonUtil.BuildJsonSuccess(nil)
	}
	a, _ := json.Marshal(successJson)
	response.Write(a)
}

func (d *DefaultHttpViewRender) ErrorRender(local *context.LocalStack, controller *ControllerVar,
	proxyMethod *proxyclass.ProxyMethod, methodRequestSetting *RestAnnotationSetting,
	request *http.Request, response http.ResponseWriter, throwable interface{}) {

	response.Header().Add("Content-Type", "application/json;charset=UTF-8")

	var errJson *vo.JsonResult
	switch throwable.(type) {
	case *exception.FrameException:
		value, _ := throwable.(*exception.FrameException)
		errJson = util.JsonUtil.BuildJsonFailure(value.Code, value.Message, nil)
	case error:
		err := throwable.(error)
		tip := err.Error()
		errJson = util.JsonUtil.BuildJsonFailure1(tip, nil)
	default:
		tip := fmt.Sprintln(throwable)
		errJson = util.JsonUtil.BuildJsonFailure1(tip, nil)
	}
	a, _ := json.Marshal(errJson)
	response.Write(a)
}

var defaultHttpViewRender DefaultHttpViewRender = DefaultHttpViewRender{}

var DefaultServConfig *ServerConfig

type RestAnnotationSetting struct {

	//对应path路径 以/开头
	HttpPath string

	//http method get,post,put,delete,*
	HttpMethod string

	// 方法对应的request参数名
	QueryParameter string

	// 路径参数
	PathVariable string

	// header 参数
	HeaderParameter string

	//默认的渲染类型 json html 默认是json
	MethodRender string
}

type ControllerVar struct {
	Target               proxyclass.ProxyTarger
	PrefixPath           string
	AbsoluteMethodPath   map[string]*proxyclass.ProxyMethod
	NoAbsoluteMethodPath []*proxyclass.ProxyMethod
	NoAbsolutePathTree   *PathNode
}

type DispatchServlet struct {
	routeMapping []*RouteMapping
	defaultView  HttpViewRender
	logger       logclass.AppLoger
	lock         sync.Mutex
}

func (d *DispatchServlet) initDispatchServlet(local *context.LocalStack,
	applicationContext *application.FrameApplicationContext) {
	if d.logger != nil {
		return
	}
	d.lock.Lock()
	defer func() {
		d.lock.Unlock()
	}()
	if d.logger != nil {
		return
	}
	d.logger = applicationContext.LogFactory.GetLoggerType(reflect.TypeOf(d).Elem())
}

func (d *DispatchServlet) SetDefaultView(view HttpViewRender) {
	d.defaultView = view
}

func (d *DispatchServlet) GetRouteMapping() []*RouteMapping {
	return d.routeMapping
}

func (d *DispatchServlet) Dispatch(local *context.LocalStack, request *http.Request, response http.ResponseWriter) {
	controller := GetCurrentControllerInvoker(local)

	var proxyMethod *proxyclass.ProxyMethod
	var methodRequestSetting *RestAnnotationSetting

	var returnError interface{}
	defer func() {
		if err := recover(); err != nil {
			returnError = err
		}
		if returnError != nil {
			d.httpErrorRender(local, controller, proxyMethod, methodRequestSetting, request, response, returnError)
		}
	}()

	// 去除?之后的
	url := clearHttpPath(request.URL.Path)
	url = removePrefix(url, controller.PrefixPath)

	httpMethod := strings.ToLower(request.Method)
	mk := fmt.Sprintf("%s-%s", httpMethod, url)

	// absolutepath
	if _, ok := controller.AbsoluteMethodPath[mk]; ok {
		proxyMethod = controller.AbsoluteMethodPath[mk]
	} else {
		mk = fmt.Sprintf("%s-%s", "*", url)
		proxyMethod = controller.AbsoluteMethodPath[mk]
	}

	// noabsolute path
	var pathVariableValue map[string]string = make(map[string]string)
	if proxyMethod == nil {
		//controller.NoAbsolutePathTree.PrintTree()
		node, pv := controller.NoAbsolutePathTree.MatchMethod(url)
		if node != nil {
			proxyMethod = node.ProxyMethod
			pathVariableValue = pv
		}
	}

	if d.logger.IsDebugEnable() {
		if proxyMethod == nil {
			d.logger.Error(local, nil, "请求路径 %s 控制器 %s 处理方法 %s", url,
				util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
				"nil")
		} else {
			d.logger.Debug(local, "请求路径 %s 控制器 %s 处理方法 %s", url,
				util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
				proxyMethod.Name)
		}
	}

	// proxyMethod== nil 404 TODO

	methodRequestSetting = GetRequestAnnotationSetting(proxyMethod.Annotations)

	methodInvoker := reflect.ValueOf(controller.Target).MethodByName(proxyMethod.Name)
	paramlen := methodInvoker.Type().NumIn()

	var result []reflect.Value
	if paramlen == 0 {
		result = methodInvoker.Call([]reflect.Value{})
	} else {
		var queryParameter []string
		// TODO
		var pathVariable []string
		var headerParameter []string
		if methodRequestSetting.QueryParameter != "" {
			queryParameter = strings.Split(methodRequestSetting.QueryParameter, ",")
		}
		if methodRequestSetting.HeaderParameter != "" {
			headerParameter = strings.Split(methodRequestSetting.HeaderParameter, ",")
		}
		if methodRequestSetting.PathVariable != "" {
			pathVariable = strings.Split(methodRequestSetting.PathVariable, ",")
		}

		param := make([]reflect.Value, paramlen)
		for i := 0; i < paramlen; i++ {
			pt := methodInvoker.Type().In(i)
			//if pt.Kind() == reflect.Ptr {
			//	fmt.Println("method--->", pt.Elem().Name())
			//} else {
			//	fmt.Println("method--->", pt.Name())
			//}
			if d.logger.IsDebugEnable() {
				d.logger.Debug(local, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s", url,
					util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
					proxyMethod.Name, i,
					util.ClassUtil.GetJavaClassNameByType(pt))
			}

			switch pt.Kind() {
			case reflect.Ptr:
				if pt.Elem() == reflect.TypeOf(*request) {
					param[i] = reflect.ValueOf(request)
				} else if pt.Elem() == reflect.TypeOf(*local) {
					param[i] = reflect.ValueOf(local)
				} else {
					nt := reflect.New(pt.Elem())
					ntpr := nt.Interface()

					body, err := ioutil.ReadAll(request.Body)
					if err != nil {
						d.logger.Error(local, nil, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s 获取requestbody失败", url,
							util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
							proxyMethod.Name, i,
							util.ClassUtil.GetJavaClassNameByType(pt))
						panic(fmt.Errorf("read requestbody error"))
					}
					if len(body) == 0 {
						d.logger.Error(local, nil, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s 获取requestbody失败", url,
							util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
							proxyMethod.Name, i,
							util.ClassUtil.GetJavaClassNameByType(pt))
						panic(fmt.Errorf("read requestbody empty"))
					}
					json.Unmarshal(body, ntpr)
					param[i] = reflect.ValueOf(ntpr)
				}
			case reflect.Interface:
				if reflect.TypeOf(response).Implements(pt) {
					param[i] = reflect.ValueOf(response)
				}
			case reflect.String:
				pv := getParameterValueFromRequest(request, i, queryParameter, headerParameter, pathVariable, pathVariableValue)

				if d.logger.IsDebugEnable() {
					d.logger.Debug(local, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s 值 %s", url,
						util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
						proxyMethod.Name, i,
						util.ClassUtil.GetJavaClassNameByType(pt), pv)
				}

				param[i] = reflect.ValueOf(pv)
			case reflect.Int:
				pv := getParameterValueFromRequest(request, i, queryParameter, headerParameter, pathVariable, pathVariableValue)
				if d.logger.IsDebugEnable() {
					d.logger.Debug(local, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s 值 %s", url,
						util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
						proxyMethod.Name, i,
						util.ClassUtil.GetJavaClassNameByType(pt), pv)
				}
				var pvi int = 0
				if pv != "" {
					var err error
					pvi, err = strconv.Atoi(pv)
					if err != nil {
						panic(fmt.Errorf("string2int error"))
					}
				}
				param[i] = reflect.ValueOf(pvi)
			case reflect.Int64:
				pv := getParameterValueFromRequest(request, i, queryParameter, headerParameter, pathVariable, pathVariableValue)
				if d.logger.IsDebugEnable() {
					d.logger.Debug(local, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s 值 %s", url,
						util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
						proxyMethod.Name, i,
						util.ClassUtil.GetJavaClassNameByType(pt), pv)
				}
				var pvi int64 = 0
				if pv != "" {
					var err error
					pvi, err = strconv.ParseInt(pv, 10, 64)
					if err != nil {
						panic(fmt.Errorf("string2int error"))
					}
				}
				param[i] = reflect.ValueOf(pvi)
			case reflect.Struct:
				d.logger.Error(local, nil, "请求路径 %s 控制器 %s 处理方法 %s 第 %d 参数类型 %s 参数类型只支持ptr不支持struct", url,
					util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(controller.Target).Elem().Type()),
					proxyMethod.Name, i,
					util.ClassUtil.GetJavaClassNameByType(pt))
				panic(fmt.Errorf("struct only ptr"))
			}
		}
		result = methodInvoker.Call(param)
	}

	// 只能返回两个参数
	var returnObject interface{}
	if len(result) > 0 {
		i := len(result) - 1
		for ; i >= 0; i-- {
			row := result[i]
			if row.IsZero() {
				continue
			}
			if row.Type().Implements(util.FrameErrorType) {
				returnError = row.Interface()
			} else {
				returnObject = row.Interface()
			}
		}
	}
	if returnError == nil {
		d.httpSuccessRender(local, controller, proxyMethod, methodRequestSetting, request, response, returnObject)
	}

}

// httpSuccessRender result 存在为空情况
func (d *DispatchServlet) httpSuccessRender(local *context.LocalStack, controller *ControllerVar,
	proxyMethod *proxyclass.ProxyMethod,
	methodRequestSetting *RestAnnotationSetting,
	request *http.Request, response http.ResponseWriter, result interface{}) {
	d.defaultView.SuccessRender(local, controller, proxyMethod, methodRequestSetting, request, response, result)
}

func (d *DispatchServlet) httpErrorRender(local *context.LocalStack,
	controller *ControllerVar,
	proxyMethod *proxyclass.ProxyMethod,
	methodRequestSetting *RestAnnotationSetting, request *http.Request,
	response http.ResponseWriter, throwable interface{}) {
	d.defaultView.ErrorRender(local, controller, proxyMethod, methodRequestSetting, request, response, throwable)
}

var dispatchServlet DispatchServlet = DispatchServlet{
	defaultView: HttpViewRender(&defaultHttpViewRender),
}

func GetDispatchServlet() *DispatchServlet {
	return &dispatchServlet
}

type FrameHttpFactory struct {
	logger    logclass.AppLoger
	serConfig *ServerConfig
	lock      sync.Mutex
}

func (a *FrameHttpFactory) initFrameHttpFactory(local *context.LocalStack,
	applicationContext *application.FrameApplicationContext) {
	if a.logger != nil {
		return
	}
	a.lock.Lock()
	defer func() {
		a.lock.Unlock()
	}()
	if a.logger != nil {
		return
	}
	config := &ServerConfig{}
	applicationContext.Environment.GetObjectValue("server", config)
	a.serConfig = config
	a.logger = applicationContext.LogFactory.GetLoggerType(reflect.TypeOf(a).Elem())

	a.logger.Debug(local, "[初始化] http配置 %s", util.JsonUtil.To2String(a.serConfig))
	DefaultServConfig = a.serConfig
}

// AddControllerProxyTarget 思路是根据path前缀匹配到controller，在根据path和method去匹配controller具体的method
func (a *FrameHttpFactory) AddControllerProxyTarget(local1 *context.LocalStack, target1 proxyclass.ProxyTarger,
	applicationContext *application.FrameApplicationContext) {
	a.initFrameHttpFactory(local1, applicationContext)
	GetDispatchServlet().initDispatchServlet(local1, applicationContext)

	core.AddClassProxy(target1)

	var absoluteMethodPath = make(map[string]*proxyclass.ProxyMethod)
	var noAbsoluteMethodPath []*proxyclass.ProxyMethod
	var controllerRoot *PathNode = &PathNode{}
	for _, method := range target1.ProxyTarget().Methods {
		methodRestSetting := GetRequestAnnotationSetting(method.Annotations)
		if methodRestSetting == nil {
			continue
		}

		//http method
		var hm = strings.ToLower(methodRestSetting.HttpMethod)
		if hm == "" {
			hm = "*"
		}

		var hp = methodRestSetting.HttpPath
		if hp == "/" {
			hp = ""
		}

		mkey := fmt.Sprintf("%s-%s", hm, hp)
		if strings.Index(hp, "{") >= 0 || strings.Index(hp, "*") >= 0 {
			noAbsoluteMethodPath = append(noAbsoluteMethodPath, method)
			controllerRoot.SetPath(hp, method)
			a.logger.Debug(local1, "控制器 %s 正则路径设置 %s 对应实现方法 %s",
				util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(target1).Elem().Type()),
				mkey, method.Name) // controllerRoot.PrintTree()
		} else {
			absoluteMethodPath[mkey] = method
			a.logger.Debug(local1, "控制器 %s 绝对路径设置 %s 对应实现方法 %s",
				util.ClassUtil.GetJavaClassNameByType(reflect.ValueOf(target1).Elem().Type()),
				mkey, method.Name)
		}
	}
	var prefix = GetControllerPathPrefix(&dispatchServlet, target1)
	invoker := &ControllerVar{
		Target:               target1,
		PrefixPath:           prefix,
		AbsoluteMethodPath:   absoluteMethodPath,
		NoAbsoluteMethodPath: noAbsoluteMethodPath,
		NoAbsolutePathTree:   controllerRoot,
	}

	f := func(invoker1 *ControllerVar) func(http.ResponseWriter, *http.Request) {
		return func(response http.ResponseWriter, request *http.Request) {
			local := context.NewLocalStack()
			local.SetThread()
			SetCurrentControllerInvoker(local, invoker1)

			SetCurrentHttpRequest(local, request)
			SetCurrentHttpResponse(local, response)

			defer local.Destroy()

			GetDefaultFilterChain().DoFilter(local, request, response)
		}
	}(invoker)
	var f1 = &RouteMapping{
		Path:         strings.TrimSpace(prefix + "/"),
		Handler:      f,
		AbsolutePath: false,
		Invoker:      invoker,
	}
	var f2 = &RouteMapping{
		Path:         strings.TrimSpace(prefix),
		Handler:      f,
		AbsolutePath: true,
		Invoker:      invoker,
	}
	dispatchServlet.routeMapping = append(dispatchServlet.routeMapping, f1, f2)
}

var frameHttpFactory FrameHttpFactory = FrameHttpFactory{}

func GetFrameHttpFactory() *FrameHttpFactory {
	return &frameHttpFactory
}

// getParameterValueFromRequest 获取方法常规变量
func getParameterValueFromRequest(request *http.Request, methodPosition int,
	queryParameter []string,
	headerParameter []string,
	pathVariable []string,
	pathVariableValue map[string]string,
) string {
	if len(queryParameter) > methodPosition && queryParameter[methodPosition] != "" && queryParameter[methodPosition] != "_" {
		//query parameter
		var pk string = queryParameter[methodPosition]
		var pv string = request.FormValue(pk)
		if pv == "" {
			pv = request.URL.Query().Get(pk)
		}
		return pv
	}
	if len(headerParameter) > methodPosition && headerParameter[methodPosition] != "" && headerParameter[methodPosition] != "_" {
		var pk string = headerParameter[methodPosition]
		var pv string = request.Header.Get(pk)
		return pv
	}
	if len(pathVariable) > methodPosition && pathVariable[methodPosition] != "" && pathVariable[methodPosition] != "_" {
		var pk string = pathVariable[methodPosition]
		if pv, ok := pathVariableValue[pk]; ok {
			return pv
		} else {
			return ""
		}
	}
	return ""
}

func clearHttpPath(path string) string {
	return path
}

func removePrefix(path string, prefix string) string {

	//	fmt.Println(path,prefix)
	if path == prefix {
		return ""
	}

	if strings.HasPrefix(path, prefix) {
		r := path[len(prefix):]
		if r == "/" {
			return ""
		}
		if r[0:1] != "/" {
			return fmt.Sprintf("/%s", r)
		}
		return r
	}
	return path
}

func GetControllerPathPrefix(dispatchServlet *DispatchServlet, target proxyclass.ProxyTarger) string {
	//context-path
	var sp string = ""
	if DefaultServConfig != nil && DefaultServConfig.Servlet != nil {
		sp = DefaultServConfig.Servlet.ContextPath
	}
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

type WebServletStartedEvent struct {
}

func (w *WebServletStartedEvent) GetSource() interface{} {
	return nil
}

func AddHttpViewRender(view HttpViewRender) {
	core.AddClassProxy(view.(proxyclass.ProxyTarger))
	GetDispatchServlet().SetDefaultView(view)
}
