package application

import (
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

var LoadInstanceHandlerType reflect.Type = reflect.Zero(reflect.TypeOf((*LoadInstanceHandler)(nil)).Elem()).Type()
var LoadInstanceHandlerTypeName string = util.ClassUtil.GetClassNameByType(LoadInstanceHandlerType)

var AppLogerType reflect.Type = reflect.Zero(reflect.TypeOf((*AppLoger)(nil)).Elem()).Type()
var AppLogerTypeName string = util.ClassUtil.GetClassNameByType(AppLogerType)

var AnnotationSpiType reflect.Type = reflect.Zero(reflect.TypeOf((*AnnotationSpi)(nil)).Elem()).Type()
var AnnotationSpiTypeName string = util.ClassUtil.GetClassNameByType(AnnotationSpiType)
