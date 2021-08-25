package application

import (
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

//FrameApplicationContext 应用上下文
type ApplicationContext struct {

	//应用配置和环境变量
	AppConfig *ApplicationConfig

	//?
	ValueBindTree *InsValueInjectTree

	//适配map 加载的实例如果实现ProxyInstanceAdapter接口的话 取出key
	//AdapterMap    map[string]map[string]*DynamicProxyInstanceNode

	// 资源池
	FrameResource *ResourcePool

	// 自定义http启动器
	FrameHttpStarter HttpStarter
}

func (f *ApplicationContext) ProxyTarget() *proxyclass.ProxyClass {
	return nil
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

func (f *ApplicationContext) GetProxyInsById(id string) proxyclass.ProxyTarger {
	if result, ok := f.FrameResource.ProxyInsPool.ElementMap[id]; !ok {
		return nil
	} else {
		return result.Target
	}
}
