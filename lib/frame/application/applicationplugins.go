package application

import (
	"context"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type HttpStarter interface {
	HttpStart(local context.Context, applicationContext *ApplicationContext)
}

//type ProxyInstanceAdapter interface {
//	// AdapterKey 返回长度1-2个
//	AdapterKey() []string
//}

/**
应用启动事件监听器
*/
type ApplicationContextListener interface {
	Starting(local context.Context)

	EnvironmentPrepared(local context.Context, appConfig *ApplicationConfig)

	Running(local context.Context, applicationContext *ApplicationContext)

	Failed(local context.Context, applicationContext *ApplicationContext, err interface{})

	Order() int
}

/**
启动事件统一处理
*/
type ApplicationRunContextListeners struct {
	ApplicationListeners []ApplicationContextListener
	Args                 *ApplicationArguments
}

func (a *ApplicationRunContextListeners) Starting(local context.Context) {
	for _, m := range a.ApplicationListeners {
		m.Starting(local)
	}
}
func (a *ApplicationRunContextListeners) EnvironmentPrepared(local context.Context, environment *ApplicationConfig) {
	for _, m := range a.ApplicationListeners {
		m.EnvironmentPrepared(local, environment)
	}
}
func (a *ApplicationRunContextListeners) Running(local context.Context, applicationContext *ApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.Running(local, applicationContext)
	}
}
func (a *ApplicationRunContextListeners) Failed(local context.Context, applicationContext *ApplicationContext, err interface{}) {
	for _, m := range a.ApplicationListeners {
		m.Failed(local, applicationContext, err)
	}
}

// LoadInstanceHandler 加载实例的时候
/**
第一轮 实例字段注入完成之后 检查实例是否需要特殊处理
例如factory实例会生成其他实例
*/
type LoadInstanceHandler interface {

	//LoadInstance 返回的bool true 该handler已经处理 false 不处理
	LoadInstance(local context.Context, target *DynamicProxyInstanceNode,
		application *Application,
		applicationContext *ApplicationContext) (bool, error)

	Order() int
}

/**
双向链表 用来保存容器实例
*/
type DynamicProxyLinkedArray struct {
	// 首节点
	FirstElement *DynamicProxyInstanceNode

	// 实例pool id对应的实例 util.ClassUtil.GetSimpleClassName
	ElementMap map[string]*DynamicProxyInstanceNode

	// Element类型名 对应的类型 util.ClassUtil.GetClassName
	ElementTypeNameMap map[string]*DynamicProxyInstanceNode

	// 要注入的接口类型 util.ClassUtil.GetClassName
	InterfaceTypeNameMap map[string]reflect.Type

	// 平台的接口 util.ClassUtil.GetClassName
	SysInterfaceTypeNameMap map[string]reflect.Type
	// 注入实例的时候检查实例是否属于平台接口
	SysInterfaceImplNameMap map[string][]*DynamicProxyInstanceNode

	LogFactory AppLogFactoryer

	// 最后一个节点
	LastElement *DynamicProxyInstanceNode

	initLock sync.Once
}

func (d *DynamicProxyLinkedArray) init() {
	d.initLock.Do(func() {
		if d.ElementMap == nil {
			d.ElementMap = make(map[string]*DynamicProxyInstanceNode)
		}
		if d.ElementTypeNameMap == nil {
			d.ElementTypeNameMap = make(map[string]*DynamicProxyInstanceNode)
		}
		if d.InterfaceTypeNameMap == nil {
			d.InterfaceTypeNameMap = make(map[string]reflect.Type)
		}
		if d.SysInterfaceTypeNameMap == nil {
			d.SysInterfaceTypeNameMap = make(map[string]reflect.Type)
		}
		if d.SysInterfaceImplNameMap == nil {
			d.SysInterfaceImplNameMap = make(map[string][]*DynamicProxyInstanceNode)
		}
	})
}

/*
AddHead
往头部加入元素
当一些节点解析优先级比较高的时候 在放入容器的时候就往后放
*/
func (d *DynamicProxyLinkedArray) AddHead(node *DynamicProxyInstanceNode) {
	d.init()

	if _, ok := d.ElementMap[node.Id]; ok {
		err := fmt.Errorf("%s already add，please rename it", node.Id)
		panic(err)
	}

	if node.Target != nil {
		target := node.Target
		node.rt = reflect.TypeOf(target)
		node.rv = reflect.ValueOf(target)
		fieldNum := node.rt.Elem().NumField()
		if fieldNum > 0 {
			for i := 0; i < fieldNum; i++ {
				field := node.rt.Elem().Field(i)
				if _, ok := field.Tag.Lookup(AutowiredInjectKey); ok {
					node.instanceInject = append(node.instanceInject, &field)
					if field.Type.Kind() == reflect.Interface {
						d.AddInterfacer(field.Type)
					}
				}

				if _, ok := field.Tag.Lookup(ValueInjectKey); ok {
					node.configInjectField = append(node.configInjectField, &field)
				}
			}
		}
	}

	if d.LastElement == nil {
		d.LastElement = node
	}

	if d.FirstElement == nil {
		d.FirstElement = node
	} else {
		oldFirst := d.FirstElement
		node.Next = oldFirst
		d.FirstElement = node
	}

	d.ElementMap[node.Id] = node
}

func (d *DynamicProxyLinkedArray) Push(node *DynamicProxyInstanceNode) {
	d.init()

	if _, ok := d.ElementMap[node.Id]; ok {
		err := fmt.Errorf("%s already add，please rename it", node.Id)
		panic(err)
	}

	if node.Target != nil {
		target := node.Target
		node.rv = reflect.ValueOf(target)
		node.rt = reflect.ValueOf(target).Type()
		fieldNum := node.rt.Elem().NumField()
		if fieldNum > 0 {
			for i := 0; i < fieldNum; i++ {
				field := node.rt.Elem().Field(i)
				if _, ok := field.Tag.Lookup(AutowiredInjectKey); ok {
					node.instanceInject = append(node.instanceInject, &field)
					if field.Type.Kind() == reflect.Interface {
						d.AddInterfacer(field.Type)
					}
				}

				if _, ok := field.Tag.Lookup(ValueInjectKey); ok {
					node.configInjectField = append(node.configInjectField, &field)
				}
			}
		}
	}

	if d.FirstElement == nil {
		d.FirstElement = node
	}

	d.ElementMap[node.Id] = node

	if d.LastElement != nil {
		d.LastElement.Next = node
	}
	d.LastElement = node
}

// 这里只接收interface类型
func (d *DynamicProxyLinkedArray) AddInterfacer(t reflect.Type) {
	d.init()
	if t.Kind() != reflect.Interface {
		err := fmt.Errorf("%s is not interface type", t.Name())
		panic(err)
	}
	d.InterfaceTypeNameMap[util.ClassUtil.GetClassNameByTypeV1(t)] = t
}

func (d *DynamicProxyLinkedArray) AddSysInterfacer(t reflect.Type) {
	d.init()
	if t.Kind() != reflect.Interface {
		err := fmt.Errorf("%s is not interface type", t.Name())
		panic(err)
	}
	name := util.ClassUtil.GetClassNameByTypeV1(t)
	d.InterfaceTypeNameMap[name] = t
	d.SysInterfaceTypeNameMap[name] = t
}

/*
DynamicProxyInstanceNode
实例节点
*/
type DynamicProxyInstanceNode struct {

	// 目标对象 指针
	Target interface{}

	ClassInfo *ClassV1

	// 下一个节点
	Next *DynamicProxyInstanceNode

	// 节点id 如果不指定到话  see util.ClassUtil.GetSimpleClassName
	Id string

	// Target 类型 push的时候设置  里面的值都是指针
	rt reflect.Type

	// Target 反射值 push的时候设置 里面的值都是指针
	rv reflect.Value

	// 需要注入配置的字段 push的时候设置
	configInjectField []*reflect.StructField

	// 需要注入实例的字段 push的时候设置
	instanceInject []*reflect.StructField
}

/*
InsValueInjectTree
需要注入的配置项，用于缓存
例如两个地方a.b都注入同一个实例Aa那么两处使用同一个对象
*/
type InsValueInjectTree struct {
	Root      *InsValueInjectTreeNode
	RefNode   map[string]*InsValueInjectTreeNode
	AppConfig *ApplicationConfig
}

func (i *InsValueInjectTree) SetTreeNode(key string,
	baseVal string, //基础类型值
	objectVal interface{}, //对象值 指针
	node *DynamicProxyInstanceNode,
	field *reflect.StructField,
	defaultVal string) {
	if i.Root == nil {
		i.Root = &InsValueInjectTreeNode{
			ChildrenMap: make(map[string]*InsValueInjectTreeNode),
			Children:    make([]*InsValueInjectTreeNode, 0, 5),
		}
	}
	keys := strings.Split(key, ".")

	var current *InsValueInjectTreeNode = i.Root
	for n := 0; n < len(keys); n++ {
		k := keys[n]
		if children, ok := current.ChildrenMap[k]; !ok {
			child := &InsValueInjectTreeNode{
				Key:         strings.Join(keys[0:n+1], "."),
				ObjectValue: make(map[string]interface{}),
				BaseValue:   "",
				ChildrenMap: make(map[string]*InsValueInjectTreeNode),
			}
			current.Children = append(current.Children, child)
			current.ChildrenMap[k] = child
			i.RefNode[child.Key] = child
			current = child
		} else {
			current = children
		}

		if n == (len(keys) - 1) {
			current.OwnerField = append(current.OwnerField, field)
			current.OwnerTarget = append(current.OwnerTarget, node)
			current.BaseValue = baseVal
			current.DefaultValue = defaultVal
			switch field.Type.Kind() {
			case reflect.Map:
				// 先不管吧
				current.MapValue = append(current.MapValue, objectVal)
			case reflect.Ptr:
				name := util.ClassUtil.GetClassNameByTypeV1(field.Type.Elem())
				current.ObjectValue[name] = objectVal
			case reflect.Struct:
				name := util.ClassUtil.GetClassNameByTypeV1(field.Type)
				current.ObjectValue[name] = objectVal
			}
		}
	}

}

// getValueForType 只有struct ptr-struct base缓存（每次设置的时候检查之前有没有生成过，如果有的话就用之前的）
// map结构每次都重新生成
func (i *InsValueInjectTree) getValueForType(key string, t reflect.Type) *reflect.Value {
	var node *InsValueInjectTreeNode
	var ok = false

	if node, ok = i.RefNode[key]; !ok {
		return nil
	}
	switch t.Kind() {
	case reflect.Map:
		return nil
	case reflect.Ptr:
		name := util.ClassUtil.GetClassNameByTypeV1(t.Elem())

		if v, ok1 := node.ObjectValue[name]; ok1 {
			m := reflect.ValueOf(v)
			return &m
		} else {
			return nil
		}
	case reflect.Struct:
		name := util.ClassUtil.GetClassNameByTypeV1(t)
		if v, ok1 := node.ObjectValue[name]; ok1 {
			m := reflect.ValueOf(v).Elem()
			return &m
		} else {
			return nil
		}
	default:
		v1 := node.BaseValue
		switch t.Kind() {
		case reflect.String:
			v2 := reflect.ValueOf(v1)
			return &v2
		case reflect.Int64:
			if v1 == "" {
				v1 = "0"
			}
			v2, _ := strconv.ParseInt(v1, 10, 64)
			val1 := reflect.ValueOf(v2)
			return &val1
		case reflect.Int:
			if v1 == "" {
				v1 = "0"
			}
			v2, _ := strconv.Atoi(v1)
			val1 := reflect.ValueOf(v2)
			return &val1
		case reflect.Float64:
			if v1 == "" {
				v1 = "0.0"
			}
			v2, _ := strconv.ParseFloat(v1, 64)
			val1 := reflect.ValueOf(v2)
			return &val1
		}
	}
	return nil
}

// SetBindValue configkey中不支持数组
func (i *InsValueInjectTree) SetBindValue(
	target *DynamicProxyInstanceNode,
	field *reflect.StructField,
	configkey string,
	defaultVal string,
) {
	var val *reflect.Value

	if configkey != "" {
		val = i.getValueForType(configkey, field.Type)
		if val == nil {
			var baseVal string
			var objectVal interface{} // ptr
			switch field.Type.Kind() {
			case reflect.Map:
				//valvalrt := valrt.Elem()
				valmaprt := reflect.MapOf(reflect.TypeOf(""), field.Type.Elem())
				valmaprv := reflect.MakeMap(valmaprt)
				i.AppConfig.GetObjectValue(configkey, valmaprv.Interface())
				objectVal = valmaprv.Interface()
				val = &valmaprv
			case reflect.Ptr:
				v := reflect.New(field.Type.Elem())
				i.AppConfig.GetObjectValue(configkey, v.Interface())
				objectVal = v.Interface()
				val = &v
			case reflect.Struct:
				v := reflect.New(field.Type)
				i.AppConfig.GetObjectValue(configkey, v.Interface())
				objectVal = v.Interface()
				v1 := v.Elem()
				val = &v1
			default:
				v1 := i.AppConfig.GetBaseValue(configkey, defaultVal)
				baseVal = v1
				switch field.Type.Kind() {
				case reflect.String:
					v2 := reflect.ValueOf(v1)
					val = &v2
				case reflect.Int64:
					if v1 != "" {
						v2, _ := strconv.ParseInt(v1, 10, 64)
						val1 := reflect.ValueOf(v2)
						val = &val1
					}
				case reflect.Int:
					if v1 != "" {
						v2, _ := strconv.Atoi(v1)
						val1 := reflect.ValueOf(v2)
						val = &val1
					}
				case reflect.Float64:
					if v1 != "" {
						v2, _ := strconv.ParseFloat(v1, 64)
						val1 := reflect.ValueOf(v2)
						val = &val1
					}
				}
			}
			//AddTreeNode
			i.SetTreeNode(configkey,
				baseVal,   //基础类型值
				objectVal, //对象值 指针
				target,
				field,
				defaultVal)
		}
	} else {
		switch field.Type.Kind() {
		case reflect.String:
			val1 := reflect.ValueOf(defaultVal)
			val = &val1
		case reflect.Int64:
			if defaultVal != "" {
				v2, _ := strconv.ParseInt(defaultVal, 10, 64)
				val1 := reflect.ValueOf(v2)
				val = &val1
			}
		case reflect.Int:
			if defaultVal != "" {
				v2, _ := strconv.Atoi(defaultVal)
				val1 := reflect.ValueOf(v2)
				val = &val1
			}
		case reflect.Float64:
			if defaultVal != "" {
				v2, _ := strconv.ParseFloat(defaultVal, 64)
				val1 := reflect.ValueOf(v2)
				val = &val1
			}
		}
	}
	if val != nil {
		target.rv.Elem().FieldByName(field.Name).Set(*val)
	}
}

type InsValueInjectTreeNode struct {
	// key 关键字
	Key string

	// 对象值都是指针struct结构
	ObjectValue map[string]interface{}

	// 对象如果是map类型
	MapValue []interface{}

	// 基础值都是string类型
	BaseValue string

	// 默认值
	DefaultValue string

	ChildrenMap map[string]*InsValueInjectTreeNode
	Children    []*InsValueInjectTreeNode

	// 有哪些字段绑定了这个关键字 target和field是一一对应的
	OwnerTarget []*DynamicProxyInstanceNode
	OwnerField  []*reflect.StructField
}

type AppLogFactoryer interface {
	GetLoggerType(p reflect.Type) AppLoger

	GetLoggerString(className string) AppLoger
}

type Annotation interface {
	AnnotationValue() interface{}
}

type AnnotationSpi interface {
	AnnotationName() string

	NewAnnotation() Annotation
}

type AppLoger interface {
	Trace(local context.Context, format string, a ...interface{})

	IsTraceEnable() bool

	Debug(local context.Context, format string, a ...interface{})

	IsDebugEnable() bool

	Info(local context.Context, format string, a ...interface{})

	IsInfoEnable() bool

	Warn(local context.Context, format string, a ...interface{})

	IsWarnEnable() bool

	//err 可空
	Error(local context.Context, err interface{}, format string, a ...interface{})

	IsErrorEnable() bool
}
