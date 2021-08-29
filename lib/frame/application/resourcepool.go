package application

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
)

// ResourcePool 资源pool 所有启动加载的资源都放这里 例如配置文件 代理单例
type ResourcePool struct {
	// 配置文件
	ConfigMap map[string]string
	// 日志配置文件
	LogConfigMap map[string]string
	// 实例池
	ProxyInsPool *DynamicProxyLinkedArray
}

// RegisterInterfaceType 接口类型 可以见RegisterInterfaceType(ApplicationContextListenerType)
// 后面可以通过applicationContext.GetProxyInsByInterfaceType 获取
// 因为go不能通过struct反推出它所包含的接口,所以先把所有接口注入到容器，在注入struct的时候和接口一个个比较
func (r *ResourcePool) RegisterInterfaceType(t reflect.Type) {
	r.ProxyInsPool.AddInterfacer(t)
}

// AddConfigYaml 初始化添加全局配置文件内容
//  name 配置关键字 ，例如 ApplicationDefaultYaml
// 加载规则，默认加载default,然后从启动项或者环境变量中获取spring.profiles.active，如果获取到就加载，获取不到就加载local
func (r *ResourcePool) AddConfigYaml(name string, config string) {
	resourcePool.ConfigMap[name] = config
}

// AddAppLogConfig 初始化添加全局配置文件内容
//  name 配置关键字 ，例如 ApplicationDefaultYaml
// 加载规则，默认加载default,然后从启动项或者环境变量中获取spring.profiles.active，如果获取到就加载，获取不到就加载local
func (r *ResourcePool) AddAppLogConfig(name string, config string) {
	resourcePool.LogConfigMap[name] = config
}

// AddInstance name 可以为空 ，默认会设置类名 将实例放到容器中
// instance 必须是指针
func (r *ResourcePool) AddInstance(name string, instance interface{}) {

	if reflect.TypeOf(instance).Kind() != reflect.Ptr {
		err := fmt.Errorf("%s is not ptr", name)
		panic(err)
	}

	var key string = name

	if key == "" {
		key = util.ClassUtil.GetSimpleClassName(instance)
	}

	node := &DynamicProxyInstanceNode{
		Target: instance,
		Id:     key,
	}

	resourcePool.ProxyInsPool.Push(node)
	//var hasAdd bool = false
	//t := reflect.TypeOf(instance)
	//for k, v := range resourcePool.RegisterType {
	//	if t.Implements(v) {
	//		if as, ok := resourcePool.RegisterInsMap[k]; ok {
	//			resourcePool.RegisterInsMap[k] = append(as, node)
	//		} else {
	//			resourcePool.RegisterInsMap[k] = []*DynamicProxyInstanceNode{node}
	//		}
	//		if !hasAdd {
	//			resourcePool.ProxyInsPool.AddHead(node)
	//			hasAdd = true
	//		}
	//	}
	//}
	//if !hasAdd {
	//	resourcePool.ProxyInsPool.Push(node)
	//}
}

var resourcePool ResourcePool = ResourcePool{
	ConfigMap:    make(map[string]string),
	LogConfigMap: make(map[string]string),
	ProxyInsPool: &DynamicProxyLinkedArray{},
}

func GetResourcePool() *ResourcePool {
	return &resourcePool
}

func init() {
	// 默认添加
	resourcePool.RegisterInterfaceType(ApplicationContextListenerType)
	resourcePool.RegisterInterfaceType(LoadInstanceHandlerType)
	//resourcePool.RegisterInterfaceType(FrameLogFactoryerType)
}
