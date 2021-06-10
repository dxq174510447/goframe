package application

import (
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
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
	// 注入的时候需要检查是否实现的接口
	RegisterType map[string]reflect.Type
	// 如果满足上面条件就直接放到一个list分组中
	RegisterInsMap map[string][]*DynamicProxyInstanceNode
}

func (r *ResourcePool) RegisterInterfaceType(t reflect.Type) {
	clname := util.ClassUtil.GetClassNameByType(t)
	r.RegisterType[clname] = t
}

var resourcePool ResourcePool = ResourcePool{
	ConfigMap:      make(map[string]string),
	LogConfigMap:   make(map[string]string),
	ProxyInsPool:   &DynamicProxyLinkedArray{},
	RegisterInsMap: make(map[string][]*DynamicProxyInstanceNode),
	RegisterType:   make(map[string]reflect.Type),
}

func GetResourcePool() *ResourcePool {
	return &resourcePool
}

// AddConfigYaml 初始化添加全局配置文件内容
//  name 配置关键字 ，例如 ApplicationDefaultYaml
// 加载规则，默认加载default,然后从启动项或者环境变量中获取spring.profiles.active，如果获取到就加载，获取不到就加载local
func AddConfigYaml(name string, config string) {
	resourcePool.ConfigMap[name] = config
}

// AddAppLogConfig 初始化添加全局配置文件内容
//  name 配置关键字 ，例如 ApplicationDefaultYaml
// 加载规则，默认加载default,然后从启动项或者环境变量中获取spring.profiles.active，如果获取到就加载，获取不到就加载local
func AddAppLogConfig(name string, config string) {
	resourcePool.LogConfigMap[name] = config
}

// RegisterInterfaceType 接口类型 可以见RegisterInterfaceType(ApplicationContextListenerType)
// 后面可以通过applicationContext.GetProxyInsByInterfaceType 获取
func RegisterInterfaceType(t reflect.Type) {
	resourcePool.RegisterInterfaceType(t)
}

// AddProxyInstance name 可以为空 ，默认会设置类名 将实例放到容器中
func AddProxyInstance(name string, instance proxyclass.ProxyTarger) {
	var key string = name

	if key == "" && instance.ProxyTarget() != nil && instance.ProxyTarget().Name != "" {
		key = instance.ProxyTarget().Name
	}

	if key == "" {
		key = util.ClassUtil.GetClassName(instance)
	}

	node := &DynamicProxyInstanceNode{
		Target: instance,
		Id:     key,
	}
	resourcePool.ProxyInsPool.Push(node)

	t := reflect.TypeOf(instance)
	for k, v := range resourcePool.RegisterType {
		if t.Implements(v) {
			if as, ok := resourcePool.RegisterInsMap[k]; ok {
				resourcePool.RegisterInsMap[k] = append(as, node)
			} else {
				resourcePool.RegisterInsMap[k] = []*DynamicProxyInstanceNode{node}
			}
		}
	}
}

func init() {
	// 默认添加
	resourcePool.RegisterInterfaceType(ApplicationContextListenerType)
	resourcePool.RegisterInterfaceType(FrameLoadInstanceHandlerType)
	resourcePool.RegisterInterfaceType(FrameLogFactoryerType)
}
