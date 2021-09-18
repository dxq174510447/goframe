package application

import (
	"context"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/ctx"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"gopkg.in/yaml.v3"
	"reflect"
	"sort"
	"strings"
)

type Application struct {
	// 启动类
	MainClass interface{}

	// 应用上下文加载监听器
	ApplicationListeners []ApplicationContextListener

	// 类加载策略
	LoadStrategy []LoadInstanceHandler

	// 资源池 所有启动需要加载到容器中的资源
	FrameResource *ResourcePool

	// 自定义http启动器
	FrameHttpStarter HttpStarter

	logger AppLoger
}

func (a *Application) AddApplicationContextListener(listener ApplicationContextListener) *Application {
	a.ApplicationListeners = append(a.ApplicationListeners, listener)
	return a
}

func (a *Application) HttpStarter(starter HttpStarter) *Application {
	a.FrameHttpStarter = starter
	return a
}

func (a *Application) Run(args []string) *ApplicationContext {

	local := ctx.WithNewCtx(context.Background())
	util.ThreadUtil.SetThread(local, "thread-starter")

	var listeners *ApplicationRunContextListeners
	var applicationContext *ApplicationContext
	defer func() {
		if err := recover(); err != nil {
			listeners.Failed(local, applicationContext, err)
			ctx.Destory(local)
			panic(err)
		}
	}()

	// 启动参数
	appArg := NewApplicationArguments()
	// 解析启动参数
	appArg.Parse(args)

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

	// 环境变量参数解析
	appConfig := a.PrepareEnvironment(local, listeners, appArg)

	applicationContext = a.CreateApplicationContext(local, appConfig)

	a.LoadClassInfo(local, applicationContext)

	a.RefreshContext(local, applicationContext)

	listeners.Running(local, applicationContext)

	return applicationContext
}

// PrepareEnvironment 加载应用配置项
func (a *Application) PrepareEnvironment(local context.Context,
	listeners *ApplicationRunContextListeners,
	appArgs *ApplicationArguments) *ApplicationConfig {

	appConfig := NewApplicationConfig(appArgs)

	// 记载启动配置文件
	a.ConfigureEnvironment(local, appConfig, appArgs)

	listeners.EnvironmentPrepared(local, appConfig)

	appConfig.RefreshConfigTree()

	return appConfig
}

func (a *Application) CreateApplicationContext(local context.Context, appConfig *ApplicationConfig) *ApplicationContext {
	applicationContext := NewApplicationContext(appConfig, a)
	return applicationContext
}

func (a *Application) beanInject(local context.Context, node *DynamicProxyInstanceNode) {

}

// RefreshContext 加载实例
func (a *Application) RefreshContext(local context.Context, applicationContext *ApplicationContext) {

	// 容器中所有实例 采用的是链表
	insPool := a.FrameResource.ProxyInsPool
	// 实例加载器
	if len(a.LoadStrategy) > 1 {
		sort.Slice(a.LoadStrategy, func(i, j int) bool {
			return a.LoadStrategy[i].Order() < a.LoadStrategy[j].Order()
		})
	}

	current := insPool.FirstElement

	// 第一轮注入的时候 找不到对象 可能是factoryer还没生成 放到第二轮注入 暂时只支持2轮
	firstNilInject := make([]*DynamicProxyInstanceNode, 0, 30)

	for current != nil {

		// 放入到应用上下文中
		applicationContext.addInstance(current)
		for _, interType := range insPool.InterfaceTypeNameMap {
			if current.rt.Implements(interType) {
				applicationContext.addInterfaceImpl(interType, current)
			}
		}

		if len(current.instanceInject) > 0 || len(current.configInjectField) > 0 {

			// 类型注入
			for _, field := range current.instanceInject {

				if field.Type == AppLogerType {
					logger := GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(current.rt)
					reflect.ValueOf(current.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(logger))
					continue
				}

				// 安类型 或者id注入
				if key, ok := field.Tag.Lookup(AutowiredInjectKey); ok {

					var inject bool = false
					if key != "" {
						// 指定id注入
						if ele, ok1 := insPool.ElementMap[key]; ok1 {
							if ele.Target != nil {
								//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", current.Id, field.Name, ele.Id)
								reflect.ValueOf(current.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
								inject = true
							}
						}
					} else {

						if field.Type.Kind() == reflect.Interface {
							// 根据接口类型注入
							ele := applicationContext.getByInterfaceType(field.Type)
							if ele != nil {
								if ele.Target != nil {
									//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", current.Id, field.Name, ele.Id)
									reflect.ValueOf(current.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
									inject = true
								}
							}
						} else {
							//指针注入
							name := util.ClassUtil.GetSimpleClassNameByTypeV1(field.Type.Elem())
							if ele, ok1 := insPool.ElementMap[name]; ok1 {
								if ele.Target != nil {
									//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", current.Id, field.Name, ele.Id)
									reflect.ValueOf(current.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
									inject = true
								}
							}
						}
					}

					// 如果没有注入 放到后期注入
					if !inject {
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
					//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入配置", current.Id, field.Name)
				}
			}
		}

		var add bool = false
		if len(a.LoadStrategy) > 0 {
			for _, strategy := range a.LoadStrategy {
				add = strategy.LoadInstance(local, current, a, applicationContext)
				if add {
					//a.logger.Debug(local, "[初始化] 自定义加载 加载器 %s 加载 %s",
					//	util.ClassUtil.GetJavaClassNameByType(reflect.TypeOf(reflect.ValueOf(strategy).Elem().Interface())), current.Id)
					break
				}
			}
		}
		//if !add {
		//	core.AddClassProxy(current.Target)
		//}

		current = current.Next
	}

	// 第二轮注入
	for _, ele1 := range firstNilInject {

		if ele1 == nil {
			continue
		}

		for _, field := range ele1.instanceInject {

			// 特殊类型 日志类型注入 第一轮已经注入
			if field.Type == AppLogerType {
				continue
			}

			if key, ok := field.Tag.Lookup(AutowiredInjectKey); ok {

				// 字段值为空的话 第二轮注入
				zer := reflect.ValueOf(ele1.Target).Elem().FieldByName(field.Name).IsZero()
				if !zer {
					continue
				}

				var inject bool = false
				if key != "" {
					if ele, ok1 := insPool.ElementMap[key]; ok1 {
						if ele.Target != nil {
							//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", ele1.Id, field.Name, ele.Id)
							reflect.ValueOf(ele1.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
							inject = true
						}
					}
				} else {
					if field.Type.Kind() == reflect.Interface {
						ele := applicationContext.getByInterfaceType(field.Type)
						if ele != nil && ele.Target != nil {
							//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", ele1.Id, field.Name, ele.Id)
							reflect.ValueOf(ele1.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
							inject = true
						}
					} else {
						//指针注入
						name := util.ClassUtil.GetSimpleClassNameByTypeV1(field.Type.Elem())
						if ele, ok1 := insPool.ElementMap[name]; ok1 {
							if ele.Target != nil {
								//a.logger.Debug(local, "[初始化] 实例加载 %s %s 注入实例id %s", ele1.Id, field.Name, ele.Id)
								reflect.ValueOf(ele1.Target).Elem().FieldByName(field.Name).Set(reflect.ValueOf(ele.Target))
								inject = true
							}
						}
					}
				}

				if !inject {
					panic(fmt.Sprintf("找不到 类型 %s 的属性 %s 的实现类",
						util.ClassUtil.GetClassNameByTypeV1(ele1.rt.Elem()),
						field.Name))
				}
			}
		}
	}

}

// ConfigureEnvironment 加载配置文件
func (a *Application) ConfigureEnvironment(local context.Context,
	environment *ApplicationConfig,
	appArgs *ApplicationArguments) {

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
		} else {
			//a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 找不到对应配置文件", f))
		}
	}

	// 加载引用文件
	fileinclude := environment.GetBaseValue("spring.profiles.include", "")
	if fileinclude != "" {
		fs := strings.Split(fileinclude, ",")
		for _, f := range fs {
			if c, ok := GetResourcePool().ConfigMap[f]; ok {
				environment.Parse(c)
				//a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 解析并加载", f))
			} else {
				//a.logs = append(a.logs, fmt.Sprintf("[初始化] 资源配置 %s 找不到对应配置文件", f))
			}
		}
	}

}

func (a *Application) LoadClassInfo(local context.Context, applicationContext *ApplicationContext) {

	if spis, ok := resourcePool.ProxyInsPool.SysInterfaceImplNameMap[AnnotationSpiTypeName]; ok {
		for _, spi := range spis {
			GetAnnotationFactory().AddAnnotationSpi(spi.Target.(AnnotationSpi))
		}
	}

	for _, classInfo := range GetResourcePool().ClassInfoList {
		var infos []*ClassV1
		err := yaml.Unmarshal([]byte(classInfo), &infos)
		if err != nil {
			panic(err)
		}
		for _, info := range infos {
			applicationContext.addClassInfo(info.PkgName, info)
		}
	}
}

func NewApplication(main interface{}) *Application {

	// 应用启动监听器
	listeners := make([]ApplicationContextListener, 0, 0)
	if arr, ok := GetResourcePool().ProxyInsPool.SysInterfaceImplNameMap[ApplicationContextListenerTypeName]; ok {
		for _, a := range arr {
			listeners = append(listeners, a.Target.(ApplicationContextListener))
		}
	}

	instanceLoad := make([]LoadInstanceHandler, 0, 0)
	if arr, ok := GetResourcePool().ProxyInsPool.SysInterfaceImplNameMap[LoadInstanceHandlerTypeName]; ok {
		for _, a := range arr {
			instanceLoad = append(instanceLoad, a.Target.(LoadInstanceHandler))
		}
	}

	app := &Application{
		MainClass:            main,
		ApplicationListeners: listeners,
		LoadStrategy:         instanceLoad,
		FrameResource:        GetResourcePool(),
	}
	app.logger = GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(reflect.TypeOf(app))
	return app
}
