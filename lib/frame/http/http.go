package http

import (
	"encoding/json"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/exception"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"github.com/dxq174510447/goframe/lib/frame/vo"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

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
}

func (d *DispatchServlet) Dispatch(local *context.LocalStack, request *http.Request, response http.ResponseWriter) {
	controller := GetCurrentControllerInvoker(local)

	var proxyMethod *proxyclass.ProxyMethod
	var methodRequestSetting *RestAnnotationSetting

	defer func() {
		if err := recover(); err != nil {
			s := PrintStackTrace(err)
			fmt.Println(s)
			if methodRequestSetting.MethodRender == "" || methodRequestSetting.MethodRender == "json" {
				d.renderExceptionJson(response, request, err)
			}

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
						panic(fmt.Errorf("read requestbody error"))
					}
					if len(body) == 0 {
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
				param[i] = reflect.ValueOf(pv)
			case reflect.Int:
				pv := getParameterValueFromRequest(request, i, queryParameter, headerParameter, pathVariable, pathVariableValue)
				var pvi int = 0
				if pv != "" {
					var err error
					pvi, err = strconv.Atoi(pv)
					if err != nil {
						panic(fmt.Errorf("string2int error"))
					}
				}
				param[i] = reflect.ValueOf(pvi)
			case reflect.Struct:
				panic(fmt.Errorf("struct only ptr"))
			}
		}
		result = methodInvoker.Call(param)
	}
	if len(result) == 1 && methodRequestSetting.MethodRender == "" || methodRequestSetting.MethodRender == "json" {
		d.renderJson(response, request, result[0].Interface())
	}

}

func (d *DispatchServlet) renderJson(response http.ResponseWriter, request *http.Request, result interface{}) {
	response.Header().Add("Content-Type", "application/json;charset=UTF-8")
	a, _ := json.Marshal(result)
	response.Write(a)
}
func (d *DispatchServlet) renderExceptionJson(response http.ResponseWriter, request *http.Request, throwable interface{}) {

	var errJson *vo.JsonResult
	switch throwable.(type) {
	case *exception.FrameException:
		value, _ := throwable.(*exception.FrameException)
		errJson = util.JsonUtil.BuildJsonFailure(value.Code, value.Message, nil)
	default:
		tip := fmt.Sprintln(throwable)
		errJson = util.JsonUtil.BuildJsonFailure1(tip, nil)
	}

	response.Header().Add("Content-Type", "application/json;charset=UTF-8")
	a, _ := json.Marshal(errJson)
	response.Write(a)
}

var dispatchServlet DispatchServlet = DispatchServlet{}

func GetDispatchServlet() *DispatchServlet {
	return &dispatchServlet
}

// AddControllerProxyTarget 思路是根据path前缀匹配到controller，在根据path和method去匹配controller具体的method
func AddControllerProxyTarget(target1 proxyclass.ProxyTarger) {
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
			// controllerRoot.PrintTree()
		} else {
			absoluteMethodPath[mkey] = method
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
			SetCurrentHttpRequest(local, request)
			SetCurrentHttpResponse(local, response)

			defer local.Destroy()

			SetCurrentControllerInvoker(local, invoker1)

			GetDefaultFilterChain().DoFilter(local, request, response)
		}
	}(invoker)
	http.HandleFunc(prefix+"/", f) //前缀匹配
	http.HandleFunc(prefix, f)     //绝对匹配
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
