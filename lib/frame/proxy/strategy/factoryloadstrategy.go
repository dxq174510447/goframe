package strategy

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
)

type ProxyFactoryer interface {

	// ProxyGet 可以自己加入实例 返回nil即可
	ProxyGet(local *context.LocalStack, application *application.FrameApplication, applicationContext *application.FrameApplicationContext) proxyclass.ProxyTarger
}

type ProxyFactoryLoadStrategy struct {
}

func (h *ProxyFactoryLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application1 *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
		return false
	}
	target := target1.Target
	if f, ok := target.(ProxyFactoryer); ok {
		core.AddClassProxy(target)
		m := f.ProxyGet(local, application1, applicationContext)
		if m != nil {
			name := util.ClassUtil.GetClassName(m)
			if m.ProxyTarget() != nil && m.ProxyTarget().Name != "" {
				name = m.ProxyTarget().Name
			}
			node1 := &application.DynamicProxyInstanceNode{
				Id:     name,
				Target: m,
			}
			application1.FrameResource.ProxyInsPool.Push(node1)
		}
		return true
	}
	return false

}

func (h *ProxyFactoryLoadStrategy) Order() int {
	return 30
}

func (h *ProxyFactoryLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var proxyFactoryLoadStrategy ProxyFactoryLoadStrategy = ProxyFactoryLoadStrategy{}

func GetProxyFactoryLoadStrategy() *ProxyFactoryLoadStrategy {
	return &proxyFactoryLoadStrategy
}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&proxyFactoryLoadStrategy))
}
