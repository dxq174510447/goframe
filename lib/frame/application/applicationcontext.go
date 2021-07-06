package application

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
	"sort"
	"strings"
)

type FrameApplicationContext struct {
	Environment   *ConfigurableEnvironment
	ValueBindTree *InsValueInjectTree
	AdapterMap    map[string]map[string]*DynamicProxyInstanceNode
	FrameResource *ResourcePool
	LogFactory    logclass.AppLogFactoryer
	// 自定义http启动器
	CustomerStarter HttpStarter
	logger          logclass.AppLoger
}

func (f *FrameApplicationContext) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

// GetProxyInsByInterfaceType t需要是接口type，并且使用RegisterInterfaceType注入过
func (f *FrameApplicationContext) GetProxyInsByInterfaceType(t reflect.Type) []proxyclass.ProxyTarger {
	name := util.ClassUtil.GetClassNameByType(t)
	targets := make([]proxyclass.ProxyTarger, 0, 0)
	if arr, ok := f.FrameResource.RegisterInsMap[name]; ok {
		for _, a := range arr {
			targets = append(targets, a.Target)
		}
	}
	return targets
}

func (f *FrameApplicationContext) GetProxyInsById(id string) proxyclass.ProxyTarger {
	if result, ok := f.FrameResource.ProxyInsPool.ElementMap[id]; !ok {
		return nil
	} else {
		return result.Target
	}
}

func (f *FrameApplicationContext) GetProxyInsByAdapterKey(groupName string, groupKey string) proxyclass.ProxyTarger {
	if _, ok := f.AdapterMap[groupName]; !ok {
		return nil
	}
	groups := f.AdapterMap[groupName]
	if result, ok := groups[groupKey]; !ok {
		return nil
	} else {
		return result.Target
	}
}

func (f *FrameApplicationContext) GetProxyInsByAdapterGroup(groupName string) []proxyclass.ProxyTarger {
	if _, ok := f.AdapterMap[groupName]; !ok {
		return nil
	}
	siz := len(f.AdapterMap[groupName])
	result := make([]proxyclass.ProxyTarger, 0, siz)

	for _, v := range f.AdapterMap[groupName] {
		result = append(result, v.Target)
	}
	return result
}

func (f *FrameApplicationContext) SetProxyInsByAdapter(groupName string, groupKey string, target *DynamicProxyInstanceNode) {
	if _, ok := f.AdapterMap[groupName]; !ok {
		f.AdapterMap[groupName] = make(map[string]*DynamicProxyInstanceNode)
	}
	f.AdapterMap[groupName][groupKey] = target
}

type FrameApplication struct {
	MainClass interface{}

	ApplicationListeners []ApplicationContextListener

	LoadStrategy []FrameLoadInstanceHandler

	Environment *ConfigurableEnvironment

	FrameResource *ResourcePool

	LogFactory logclass.AppLogFactoryer

	// 自定义http启动器
	CustomerStarter HttpStarter

	// 在日志初始化之前 记录下需要打印到日志
	// 后续是否修改成 打印日志就初始化，等资源加载完在变更日志配置 TODO 1
	logs []string

	// 只有在日志初始化之后才能使用 结合logs
	logger logclass.AppLoger
}

func (a *FrameApplication) AddApplicationContextListener(listener ApplicationContextListener) *FrameApplication {
	a.ApplicationListeners = append(a.ApplicationListeners, listener)
	return a
}

func (a *FrameApplication) HttpServ(starter HttpStarter) *FrameApplication {
	a.CustomerStarter = starter
	return a
}

func (a *FrameApplication) Run(args []string) *FrameApplicationContext {

	local := context.NewLocalStack()
	local.SetThread("")
	var listeners *ApplicationRunContextListeners
	var context1 *FrameApplicationContext
	defer func() {
		if err := recover(); err != nil {
			listeners.Failed(local, context1, err)
			local.Destroy()
			panic(err)
		}
	}()

	// 启动参数
	appArg := &DefaultApplicationArguments{
		Args:   args,
		ArgMap: make(map[string]string),
	}
	// 解析启动参数
	appArg.Parse()

	// 排序 全局启动事件回调
	if len(a.ApplicationListeners) > 1 {
		sort.Slice(a.ApplicationListeners, func(i, j int) bool {
			return a.ApplicationListeners[i].Order() < a.ApplicationListeners[j].Order()
		})
	}
	listeners = &ApplicationRunContextListeners{
		ApplicationListeners: a.ApplicationListeners,
		Args:                 appArg,
	}

	// 开始启动
	listeners.Starting(local)

	// 准备上下文环境
	a.PrepareEnvironment(local, listeners, appArg)

	// 初始化日志
	a.PrepareLogFactory(local)

	if a.logger == nil {
		a.logger = a.LogFactory.GetLoggerType(reflect.TypeOf(a).Elem())
	}

	for _, m := range a.logs {
		a.logger.Debug(local, m)
	}

	context1 = a.CreateApplicationContext(local)

	a.RefreshContext(local, context1)

	listeners.Running(local, context1)

	return context1
}

// PrepareEnvironment 加载应用配置项
func (a *FrameApplication) PrepareEnvironment(local *context.LocalStack,
	listeners *ApplicationRunContextListeners,
	appArgs *DefaultApplicationArguments) *ConfigurableEnvironment {

	c := GetEnvironmentFromApplication(local)
	if c == nil {
		c = &ConfigurableEnvironment{
			ConfigTree: &YamlTree{},
			AppArgs:    appArgs,
		}
		SetEnvironmentToApplication(local, c)

		AddProxyInstance("", proxyclass.ProxyTarger(c))
	}
	a.Environment = c

	// 记载启动配置文件
	a.ConfigureEnvironment(local, c, appArgs)
	a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 加载完成"))
	// 全局事件
	listeners.EnvironmentPrepared(local, c)

	a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 重建索引"))
	c.ConfigTree.ReIndex()

	//var m []string=make([]string,0,len(c.ConfigTree.RefNode))
	//for k,_ := range c.ConfigTree.RefNode {
	//	m = append(m,k)
	//}
	//sort.Slice(m, func(i, j int) bool {
	//	return m[i] < m[j]
	//})
	//for _,k := range m {
	//	fmt.Println(k)
	//}

	return c
}

func (a *FrameApplication) CreateApplicationContext(local *context.LocalStack) *FrameApplicationContext {
	applicationContext := &FrameApplicationContext{
		Environment: a.Environment,
		ValueBindTree: &InsValueInjectTree{
			Environment: a.Environment,
			RefNode:     make(map[string]*InsValueInjectTreeNode),
		},
		AdapterMap:      make(map[string]map[string]*DynamicProxyInstanceNode),
		FrameResource:   a.FrameResource,
		LogFactory:      a.LogFactory,
		CustomerStarter: a.CustomerStarter,
		logger:          a.LogFactory.GetLoggerType(reflect.TypeOf(a).Elem()),
	}
	AddProxyInstance("", proxyclass.ProxyTarger(applicationContext))
	return applicationContext
}

// RefreshContext 加载实例
func (a *FrameApplication) RefreshContext(local *context.LocalStack, applicationContext *FrameApplicationContext) {
	pl := GetResourcePool().ProxyInsPool

	if len(a.LoadStrategy) > 1 {
		sort.Slice(a.LoadStrategy, func(i, j int) bool {
			return a.LoadStrategy[i].Order() < a.LoadStrategy[j].Order()
		})
	}

	current := pl.FirstElement

	// 第一轮注入的时候 找不到对象 可能是factoryer还没生成 暂时只支持2轮
	firstNilInject := make([]*DynamicProxyInstanceNode, 0, 30)

	for current != nil {
		a.logger.Debug(local, "[初始化] 实例加载 %s %s 开始", current.Id, util.ClassUtil.GetJavaClassNameByType(current.rt.Elem()))
		if len(current.autowiredInjectField) > 0 || len(current.configInjectField) > 0 {
			//injectTarget = append(injectTarget, current)
			// 类型注入
			for _, field := range current.autowiredInjectField {

				// 特殊类型 日志类型注入
				if field.Type == FrameLogLoggerType {
					a.logger.Debug(local, "[初始化] 实例加载 %s %s 日志注入", current.Id, field.Name)
					if a.LogFactory != nil {
						logger := a.LogFactory.GetLoggerType(current.rt.Elem())
						reflect.ValueOf(current.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(logger))
					}
					continue
				}

				// 安类型 或者id注入
				if key, ok := field.Tag.Lookup(AutowiredInjectKey); ok {
					if key == "" {
						key = util.ClassUtil.GetClassNameByType(field.Type.Elem())
					}
					if ele, ok1 := pl.ElementMap[key]; ok1 {
						if ele.Target != nil {
							a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", current.Id, field.Name, ele.Id)
							reflect.ValueOf(current.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
						}
					} else {
						firstNilInject = append(firstNilInject, current)
					}
				}
			}

			// 配置项注入
			for _, field := range current.configInjectField {
				if key, ok := field.Tag.Lookup(ValueInjectKey); ok {
					if key == "" {
						continue
					}
					var configkey string
					var configval string
					if isContainElexpress(key) {
						b := strings.Index(key, "{")
						e := strings.LastIndex(key, "}")
						k1 := key[b+1 : e]
						d := strings.Index(k1, ":")
						if d == -1 {
							configkey = k1
						} else {
							configkey = k1[0:d]
							configval = k1[d+1 : len(k1)]
						}
						applicationContext.ValueBindTree.SetBindValue(current, field, configkey, configval)
					} else {
						applicationContext.ValueBindTree.SetBindValue(current, field, "", key)
					}
					a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入配置", current.Id, field.Name)
				}
			}
		}

		var add bool = false
		if len(a.LoadStrategy) > 0 {
			for _, strategy := range a.LoadStrategy {
				add = strategy.LoadInstance(local, current, a, applicationContext)
				if add {
					a.logger.Debug(local, "[初始化] 自定义加载 加载器 %s 加载 %s",
						util.ClassUtil.GetJavaClassNameByType(reflect.TypeOf(reflect.ValueOf(strategy).Elem().Interface())), current.Id)
					break
				}
			}
		}
		if !add {
			core.AddClassProxy(current.Target)
		}

		if f, ok := current.Target.(ProxyInstanceAdapter); ok {
			adapterKey := f.AdapterKey()
			if len(adapterKey) == 0 {
				continue
			}
			groupName := adapterKey[0]
			groupKey := current.Id
			if len(adapterKey) > 1 {
				groupKey = adapterKey[1]
			}
			applicationContext.SetProxyInsByAdapter(groupName, groupKey, current)
		}

		current = current.Next
	}

	for _, ele1 := range firstNilInject {
		if ele1 == nil {
			continue
		}
		// 类型注入
		for _, field := range ele1.autowiredInjectField {

			// 特殊类型 日志类型注入 第一轮已经注入
			if field.Type == FrameLogLoggerType {
				continue
			}

			// 安类型 或者id注入
			if key, ok := field.Tag.Lookup(AutowiredInjectKey); ok {
				// 字段值为空的话 第二轮注入
				if reflect.ValueOf(ele1.Target).Elem().FieldByName(field.Name).IsZero() {
					if key == "" {
						key = util.ClassUtil.GetClassNameByType(field.Type.Elem())
					}
					if ele, ok1 := pl.ElementMap[key]; ok1 {
						if ele.Target != nil {
							a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", ele1.Id, field.Name, ele.Id)
							reflect.ValueOf(ele1.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
						}
					} else {
						//firstNilInject = append(firstNilInject,current)
					}
				}
			}
		}
	}

}

// ConfigureEnvironment 加载配置文件
func (a *FrameApplication) ConfigureEnvironment(local *context.LocalStack,
	environment *ConfigurableEnvironment,
	appArgs *DefaultApplicationArguments) {

	// 加载系统默认配置文件
	files := make([]string, 0, 0)
	files = append(files, ApplicationDefaultYaml)
	activeFile := appArgs.GetByName("spring.profiles.active", "")
	if activeFile != "" {
		fs := strings.Split(activeFile, ",")
		files = append(files, fs...)
	} else {
		files = append(files, ApplicationLocalYaml)
	}

	for _, f := range files {
		if c, ok := GetResourcePool().ConfigMap[f]; ok {
			environment.Parse(c)
			a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 解析并加载", f))
		} else {
			a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 找不到对应配置文件", f))
		}
	}

	// 加载引用文件
	fileinclude := environment.GetBaseValue("spring.profiles.include", "")
	if fileinclude != "" {
		fs := strings.Split(fileinclude, ",")
		for _, f := range fs {
			if c, ok := GetResourcePool().ConfigMap[f]; ok {
				environment.Parse(c)
				a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 解析并加载", f))
			} else {
				a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 找不到对应配置文件", f))
			}
		}
	}

}

func (a *FrameApplication) PrepareLogFactory(local *context.LocalStack) {
	if a.LogFactory == nil {
		return
	}

	// 加载系统默认配置文件
	files := make([]string, 0, 0)
	files = append(files, ApplicationDefaultYaml)
	activeFile := a.Environment.GetBaseValue("spring.profiles.active", "")
	if activeFile != "" {
		fs := strings.Split(activeFile, ",")
		files = append(files, fs...)
	} else {
		files = append(files, ApplicationLocalYaml)
	}

	funcMap := a.Environment.GetTplFuncMap()
	for _, f := range files {
		if c, ok := GetResourcePool().LogConfigMap[f]; ok {
			a.LogFactory.Parse(c, funcMap)
			a.logs = append(a.logs, fmt.Sprintf("[初始化] 日志配置 %s 解析并加载", f))
		} else {
			a.logs = append(a.logs, fmt.Sprintf("[初始化] 日志配置 %s 找不到对应配置文件", f))
		}
	}

	a.logs = append(a.logs, fmt.Sprintf("[初始化] 日志配置 加载完成"))
}

func NewApplication(main interface{}) *FrameApplication {

	listeners := make([]ApplicationContextListener, 0, 0)
	if arr, ok := GetResourcePool().RegisterInsMap[ApplicationContextListenerTypeName]; ok {
		for _, a := range arr {
			listeners = append(listeners, a.Target.(ApplicationContextListener))
		}
	}

	instanceLoad := make([]FrameLoadInstanceHandler, 0, 0)
	if arr, ok := GetResourcePool().RegisterInsMap[FrameLoadInstanceHandlerTypeName]; ok {
		for _, a := range arr {
			instanceLoad = append(instanceLoad, a.Target.(FrameLoadInstanceHandler))
		}
	}

	var logFactory logclass.AppLogFactoryer
	if arr, ok := GetResourcePool().RegisterInsMap[FrameLogFactoryerTypeName]; ok {
		if len(arr) > 0 {
			logFactory = arr[0].Target.(logclass.AppLogFactoryer)
		}
	}

	app := &FrameApplication{
		MainClass:            main,
		ApplicationListeners: listeners,
		LoadStrategy:         instanceLoad,
		FrameResource:        GetResourcePool(),
		LogFactory:           logFactory,
	}
	return app
}
