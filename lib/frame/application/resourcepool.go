package application

import (
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"reflect"
)

// ResourcePool 资源pool 所有启动加载的资源都放这里 例如配置文件 代理单例
type ResourcePool struct {
	ConfigMap    map[string]string
	ProxyInsMap  map[string]proxy.ProxyTarger
	ProxyInsList []proxy.ProxyTarger
	// 注入的时候需要检查是否实现的接口
	RegisterType map[string]reflect.Type
	// 如果满足上面条件就直接放到一个list分组中
	RegisterInsMap map[string][]proxy.ProxyTarger
}

func (r *ResourcePool) RegisterInterfaceType(t reflect.Type) {
	clname := proxy.GetClassNameByType(t)
	r.RegisterType[clname] = t
}

var resourcePool ResourcePool = ResourcePool{
	ConfigMap:      make(map[string]string),
	ProxyInsMap:    make(map[string]proxy.ProxyTarger),
	ProxyInsList:   make([]proxy.ProxyTarger, 0, 100),
	RegisterInsMap: make(map[string][]proxy.ProxyTarger),
	RegisterType:   make(map[string]reflect.Type),
}

func GetResourcePool() *ResourcePool {
	return &resourcePool
}

func AddConfigYaml(name string, config string) {
	resourcePool.ConfigMap[name] = config
}

// AddProxyInstance name 可以为空 ，默认会设置类名
func AddProxyInstance(name string, instance proxy.ProxyTarger) {
	var key string = name
	if key == "" {
		key = proxy.GetClassName(instance)
	}

	resourcePool.ProxyInsMap[key] = instance
	resourcePool.ProxyInsList = append(resourcePool.ProxyInsList, instance)

	t := reflect.TypeOf(instance)
	for k, v := range resourcePool.RegisterType {
		if t.Implements(v) {
			if as, ok := resourcePool.RegisterInsMap[k]; ok {
				resourcePool.RegisterInsMap[k] = append(as, instance)
			} else {
				resourcePool.RegisterInsMap[k] = []proxy.ProxyTarger{instance}
			}
		}
	}
}

func init() {
	// 默认添加
	resourcePool.RegisterInterfaceType(ApplicationContextListenerType)
	resourcePool.RegisterInterfaceType(FrameLoadInstanceHandlerType)
}
