package application

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"sort"
	"strings"
)

type FrameApplicationContext struct {
	Environment *ConfigurableEnvironment
}

type FrameApplication struct {
	MainClass interface{}

	ApplicationListeners []ApplicationContextListener

	LoadStrategy []FrameLoadInstanceHandler

	Environment *ConfigurableEnvironment

	FrameResource *ResourcePool
}

func (a *FrameApplication) AddApplicationContextListener(listener ApplicationContextListener) *FrameApplication {
	a.ApplicationListeners = append(a.ApplicationListeners, listener)
	return a
}

func (a *FrameApplication) Run(args []string) *FrameApplicationContext {

	local := context.NewLocalStack()

	var listeners *ApplicationRunContextListeners
	var context *FrameApplicationContext
	defer func() {
		if err := recover(); err != nil {
			listeners.Failed(local, context, err)
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

	context = a.CreateApplicationContext(local)

	a.RefreshContext(local, context)

	listeners.Running(local, context)

	return context
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
	}
	a.Environment = c

	// 记载启动配置文件
	a.ConfigureEnvironment(local, c, appArgs)

	// 全局事件
	listeners.EnvironmentPrepared(local, c)

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
	}
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
	for current != nil {

		var add bool = false
		if len(a.LoadStrategy) > 0 {
			for _, strategy := range a.LoadStrategy {
				add = strategy.LoadInstance(local, current, a, applicationContext)
				if add {
					break
				}
			}
		}
		if !add {
			core.AddClassProxy(current.Target)
		}
		current = current.Next
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
			fmt.Printf(" 加载配置文件 %s\n", f)
			environment.Parse(c)
		} else {
			fmt.Printf(" 配置文件不存在资源中 %s\n", f)
		}
	}

	// 加载引用文件
	fileinclude := environment.GetBaseValue("spring.profiles.include", "")
	if fileinclude != "" {
		fs := strings.Split(fileinclude, ",")
		for _, f := range fs {
			if c, ok := GetResourcePool().ConfigMap[f]; ok {
				fmt.Printf(" 加载配置文件 %s\n", f)
				environment.Parse(c)
			} else {
				fmt.Printf(" 配置文件不存在资源中 %s\n", f)
			}
		}
	}

}

func NewApplication(main interface{}) *FrameApplication {

	listeners := make([]ApplicationContextListener, 0, 0)
	if arr, ok := GetResourcePool().RegisterInsMap[ApplicationContextListenerTypeName]; ok {
		for _, a := range arr {
			listeners = append(listeners, a.(ApplicationContextListener))
		}
	}

	instanceLoad := make([]FrameLoadInstanceHandler, 0, 0)
	if arr, ok := GetResourcePool().RegisterInsMap[FrameLoadInstanceHandlerTypeName]; ok {
		for _, a := range arr {
			instanceLoad = append(instanceLoad, a.(FrameLoadInstanceHandler))
		}
	}

	app := &FrameApplication{
		MainClass:            main,
		ApplicationListeners: listeners,
		LoadStrategy:         instanceLoad,
		FrameResource:        GetResourcePool(),
	}
	return app
}
