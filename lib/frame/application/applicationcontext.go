package application

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"sort"
)

type FrameApplicationContext struct {
	Environment *ConfigurableEnvironment
}

type FrameApplication struct {
	MainClass interface{}
	// 容器 所有代理带对象
	ProxyTarges []*proxy.ProxyTarger

	ApplicationListeners []ApplicationContextListener
}

func (a *FrameApplication) Run(args []string) *FrameApplicationContext {

	local := context.NewLocalStack()
	defer func() {
		defer local.Destroy()
		//TODO 错误处理
		fmt.Println("ada")
	}()

	if len(a.ApplicationListeners) > 1 {
		sort.Slice(a.ApplicationListeners, func(i, j int) bool {
			return a.ApplicationListeners[i].Order() < a.ApplicationListeners[j].Order()
		})
	}

	// 启动参数
	appArg := &DefaultApplicationArguments{Args: args[1:len(args)]}

	// 全局启动事件回调
	listeners := &ApplicationRunContextListeners{
		ApplicationListeners: a.ApplicationListeners,
		Args:                 appArg,
	}

	// 开始启动
	listeners.Starting(local)

	// 准备上下文环境
	environment := a.prepareEnvironment(local, listeners, appArg)
	fmt.Println(environment)
	//context := a.createApplicationContext(local)
	//
	//a.prepareContext(local, context, environment, listeners, appArg)
	//
	//a.refreshContext(local, context)
	//
	//listeners.Started(local, context)
	//
	//listeners.Running(local, context)

	return nil
}

// prepareEnvironment 加载应用配置项
func (a *FrameApplication) prepareEnvironment(local *context.LocalStack,
	listeners *ApplicationRunContextListeners,
	appArgs *DefaultApplicationArguments) *ConfigurableEnvironment {

	c := GetEnvironmentFromApplication(local)
	if c == nil {
		c = &ConfigurableEnvironment{PropertySources: &MutablePropertySources{}}
		SetEnvironmentToApplication(local, c)
	}

	// 记载启动配置文件
	a.configureEnvironment(local, c, appArgs)

	// 全局事件
	listeners.EnvironmentPrepared(local, c)
	return c
}

// TODO
func (a *FrameApplication) createApplicationContext(local *context.LocalStack) *FrameApplicationContext {
	return &FrameApplicationContext{}
}

// TODO
func (a *FrameApplication) prepareContext(local *context.LocalStack, context *FrameApplicationContext,
	environment *ConfigurableEnvironment,
	listeners *ApplicationRunContextListeners,
	arg *DefaultApplicationArguments) {
	context.Environment = environment

	listeners.ContextPrepared(local, context)

	a.load(local, context, a.MainClass)

	listeners.ContextLoaded(local, context)
}

func (a *FrameApplication) load(local *context.LocalStack, context *FrameApplicationContext,
	mainClass interface{}) {

}

func (a *FrameApplication) refreshContext(local *context.LocalStack, applicationContext *FrameApplicationContext) {

}

// configureEnvironment 加载配置文件
func (a *FrameApplication) configureEnvironment(local *context.LocalStack, environment *ConfigurableEnvironment, appArgs *DefaultApplicationArguments) {

}

func NewApplication(main interface{}, args ...string) *FrameApplication {
	app := &FrameApplication{MainClass: main}
	return app
}
