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
	// 实例池
	ProxyInsPool *DynamicProxyLinkedArray
	// 注入的时候需要检查是否实现的接口
	RegisterType map[string]reflect.Type
	// 如果满足上面条件就直接放到一个list分组中
	RegisterInsMap map[string][]proxyclass.ProxyTarger
}

func (r *ResourcePool) RegisterInterfaceType(t reflect.Type) {
	clname := util.ClassUtil.GetClassNameByType(t)
	r.RegisterType[clname] = t
}

var resourcePool ResourcePool = ResourcePool{
	ConfigMap:      make(map[string]string),
	ProxyInsPool:   &DynamicProxyLinkedArray{},
	RegisterInsMap: make(map[string][]proxyclass.ProxyTarger),
	RegisterType:   make(map[string]reflect.Type),
}

func GetResourcePool() *ResourcePool {
	return &resourcePool
}

func AddConfigYaml(name string, config string) {
	resourcePool.ConfigMap[name] = config
}

// AddProxyInstance name 可以为空 ，默认会设置类名
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
				resourcePool.RegisterInsMap[k] = append(as, instance)
			} else {
				resourcePool.RegisterInsMap[k] = []proxyclass.ProxyTarger{instance}
			}
		}
	}
}

func init() {
	// 默认添加
	resourcePool.RegisterInterfaceType(ApplicationContextListenerType)
	resourcePool.RegisterInterfaceType(FrameLoadInstanceHandlerType)
}
