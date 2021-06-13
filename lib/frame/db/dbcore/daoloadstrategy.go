package dbcore

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
)

type DaoLoadStrategy struct {
}

func (h *DaoLoadStrategy) LoadInstance(local *context.LocalStack, target1 *application.DynamicProxyInstanceNode,
	application *application.FrameApplication,
	applicationContext *application.FrameApplicationContext) bool {

	if target1 == nil {
		return false
	}
	target := target1.Target
	if target == nil || target.ProxyTarget() == nil {
		return false
	}
	if len(target.ProxyTarget().Annotations) == 0 {
		return false
	}
	daoAnno := GetDaoAnnotation(target.ProxyTarget().Annotations)
	if daoAnno == nil {
		return false
	}
	AddMapperProxyTarget(local, target, applicationContext)
	return true
}

func (h *DaoLoadStrategy) Order() int {
	return 10
}

func (h *DaoLoadStrategy) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var daoLoadStrategy DaoLoadStrategy = DaoLoadStrategy{}

func GetDaoLoadStrategy() *DaoLoadStrategy {
	return &daoLoadStrategy
}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&daoLoadStrategy))
}
