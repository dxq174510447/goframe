package application

import "github.com/dxq174510447/goframe/lib/frame/proxy"

// ResourcePool 资源pool 所有启动加载的资源都放这里 例如配置文件 代理单例
type ResourcePool struct {
	ConfigYaml    map[string]string
	ProxyInstance map[string]proxy.ProxyTarger
}

var resourcePool ResourcePool = ResourcePool{
	ConfigYaml:    make(map[string]string),
	ProxyInstance: make(map[string]proxy.ProxyTarger),
}

func AddConfigYaml(name string, config string) {
	resourcePool.ConfigYaml[name] = config
}

func AddProxyInstance(name string, instance proxy.ProxyTarger) {

}
