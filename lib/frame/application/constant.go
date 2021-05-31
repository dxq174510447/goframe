package application

import (
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
)

const (
	FrameEnvironmentKey = "FrameEnvironmentKey_"
)

const (
	AutowiredInjectKey = "FrameAutowired"
	ValueInjectKey     = "FrameValue"
)

const (
	ApplicationDefaultYaml = "default"
	ApplicationLocalYaml   = "local"
	ApplicationDevYaml     = "dev"
	ApplicationTestYaml    = "test"
	ApplicationUatYaml     = "uat"
	ApplicationProdYaml    = "prod"
)

var ApplicationContextListenerType reflect.Type = reflect.Zero(reflect.TypeOf((*ApplicationContextListener)(nil)).Elem()).Type()
var ApplicationContextListenerTypeName string = util.ClassUtil.GetClassNameByType(ApplicationContextListenerType)

var FrameLoadInstanceHandlerType reflect.Type = reflect.Zero(reflect.TypeOf((*FrameLoadInstanceHandler)(nil)).Elem()).Type()
var FrameLoadInstanceHandlerTypeName string = util.ClassUtil.GetClassNameByType(FrameLoadInstanceHandlerType)

func SetEnvironmentToApplication(local *context.LocalStack, env *ConfigurableEnvironment) {
	local.Set(FrameEnvironmentKey, env)
}

// GetEnvironmentFromApplication 如果上下文已经有环境上下文 就使用上一次
func GetEnvironmentFromApplication(local *context.LocalStack) *ConfigurableEnvironment {
	m := local.Get(FrameEnvironmentKey)
	if m == nil {
		return nil
	}
	return m.(*ConfigurableEnvironment)
}
