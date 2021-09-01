package application

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
)

//ApplicationContext 应用上下文
type ApplicationContext struct {

	//应用配置和环境变量
	AppConfig *ApplicationConfig

	//?
	ValueBindTree *InsValueInjectTree

	// 自定义http启动器
	FrameHttpStarter HttpStarter

	// 实例pool id对应的实例 util.ClassUtil.GetSimpleClassName
	ElementMap map[string]*DynamicProxyInstanceNode

	// Element类型名 对应的类型 util.ClassUtil.GetClassName
	// key可能是接口全名 也可能是 struct全名
	ElementTypeNameMap map[string][]*DynamicProxyInstanceNode

	logger AppLoger
}

func (a *ApplicationContext) addInstance(target *DynamicProxyInstanceNode) {
	k := target.Id
	if _, ok := a.ElementMap[k]; ok {
		err := fmt.Errorf("instace id %s alreay exists,plese rename", target.Id)
		panic(err)
	}
	a.ElementMap[k] = target

	name := util.ClassUtil.GetClassName(target.Target)
	if _, ok := a.ElementTypeNameMap[name]; ok {
		err := fmt.Errorf("instance name %s alreay exists,plese rename", name)
		panic(err)
	}
	a.ElementTypeNameMap[name] = []*DynamicProxyInstanceNode{target}
}

func (a *ApplicationContext) addInterfaceImpl(ier reflect.Type, target *DynamicProxyInstanceNode) {
	k := util.ClassUtil.GetClassNameByType(ier)
	if t, ok := a.ElementTypeNameMap[k]; ok {
		t = append(t, target)
		a.ElementTypeNameMap[k] = t
	} else {
		a.ElementTypeNameMap[k] = []*DynamicProxyInstanceNode{target}
	}
}

func (a *ApplicationContext) getByInterfaceType(ier reflect.Type) *DynamicProxyInstanceNode {
	k := util.ClassUtil.GetClassNameByType(ier)
	arr := a.ElementTypeNameMap[k]
	if len(arr) == 0 {
		return nil
	}
	return arr[0]
}

// GetProxyInsByInterfaceType t需要是接口type，并且使用RegisterInterfaceType注入过
//func (f *ApplicationContext) GetProxyInsByInterfaceType(t reflect.Type) []proxyclass.ProxyTarger {
//	name := util.ClassUtil.GetClassNameByType(t)
//	targets := make([]proxyclass.ProxyTarger, 0, 0)
//	if arr, ok := f.FrameResource.RegisterInsMap[name]; ok {
//		for _, a := range arr {
//			targets = append(targets, a.Target)
//		}
//	}
//	return targets
//}

//func (f *ApplicationContext) GetProxyInByInterfaceType(local *context.LocalStack, t reflect.Type) interface{} {
//	ts := f.GetProxyInsByInterfaceType(t)
//	if len(ts) > 0 {
//		return ts[0]
//	}
//	return nil
//}

//func (f *ApplicationContext) GetProxyInsById(id string) proxyclass.ProxyTarger {
//	if result, ok := f.FrameResource.ProxyInsPool.ElementMap[id]; !ok {
//		return nil
//	} else {
//		return result.Target
//	}
//}

func NewApplicationContext(appConfig *ApplicationConfig, application *Application) *ApplicationContext {
	applicationContext := &ApplicationContext{
		AppConfig: appConfig,
		ValueBindTree: &InsValueInjectTree{
			AppConfig: appConfig,
			RefNode:   make(map[string]*InsValueInjectTreeNode),
		},
		FrameHttpStarter:   application.FrameHttpStarter,
		ElementMap:         make(map[string]*DynamicProxyInstanceNode),
		ElementTypeNameMap: make(map[string][]*DynamicProxyInstanceNode),
	}
	applicationContext.logger = GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(reflect.TypeOf(applicationContext))
	return applicationContext
}
