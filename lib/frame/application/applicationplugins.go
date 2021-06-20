package application

import (
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type HttpStarter interface {
	HttpStart(local *context.LocalStack, applicationContext *FrameApplicationContext)
}

type ProxyInstanceAdapter interface {
	// AdapterKey 返回长度1-2个
	AdapterKey() []string
}

type ApplicationContextListener interface {
	Starting(local *context.LocalStack)

	EnvironmentPrepared(local *context.LocalStack, environment *ConfigurableEnvironment)

	Running(local *context.LocalStack, application *FrameApplicationContext)

	Failed(local *context.LocalStack, application *FrameApplicationContext, err interface{})

	Order() int
}

type ApplicationRunContextListeners struct {
	ApplicationListeners []ApplicationContextListener
	Args                 *DefaultApplicationArguments
}

func (a *ApplicationRunContextListeners) Starting(local *context.LocalStack) {
	for _, m := range a.ApplicationListeners {
		m.Starting(local)
	}
}
func (a *ApplicationRunContextListeners) EnvironmentPrepared(local *context.LocalStack, environment *ConfigurableEnvironment) {
	for _, m := range a.ApplicationListeners {
		m.EnvironmentPrepared(local, environment)
	}
}
func (a *ApplicationRunContextListeners) Running(local *context.LocalStack, application *FrameApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.Running(local, application)
	}
}
func (a *ApplicationRunContextListeners) Failed(local *context.LocalStack, application *FrameApplicationContext, err interface{}) {
	for _, m := range a.ApplicationListeners {
		m.Failed(local, application, err)
	}
}

type DefaultApplicationArguments struct {
	Args   []string
	ArgMap map[string]string
}

func (d *DefaultApplicationArguments) Parse() {

	if d.ArgMap == nil {
		d.ArgMap = make(map[string]string)
	}

	if len(d.Args) == 0 {
		return
	}

	reg := regexp.MustCompile(`^\\-+`)
	for _, arg := range d.Args {
		arg1 := reg.ReplaceAllString(strings.TrimSpace(arg), "")
		p := strings.Index(arg1, "=")
		var k1 string
		var v1 string
		if p < 0 {
			k1 = arg1
			v1 = ""
		} else {
			k1 = arg1[0:p]
			v1 = arg1[p+1 : len(arg1)]
		}
		d.ArgMap[strings.TrimSpace(k1)] = strings.TrimSpace(v1)
	}
}

func (d *DefaultApplicationArguments) GetByName(key string, defaultValue string) string {
	if m, ok := d.ArgMap[key]; ok {
		return m
	}
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	v := os.Getenv(envKey)
	if v != "" {
		return v
	}
	return defaultValue
}

type ConfigurableEnvironment struct {
	ConfigTree *YamlTree
	AppArgs    *DefaultApplicationArguments
}

func (y *ConfigurableEnvironment) GetTplFuncMap() template.FuncMap {
	return template.FuncMap{
		"env": func(key string, defaultValue string) string {
			return y.GetBaseValue(key, defaultValue)
		},
	}
}

func (y *ConfigurableEnvironment) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

func (y *ConfigurableEnvironment) Parse(content string) {
	if y.ConfigTree.AppArgs == nil {
		y.ConfigTree.AppArgs = y.AppArgs
	}
	y.ConfigTree.Parse(content)
}

func (y *ConfigurableEnvironment) GetObjectValue(key string, target interface{}) {
	y.ConfigTree.GetObjectValue(key, target)
}

func (y *ConfigurableEnvironment) GetBaseValue(key string, defaultValue string) string {
	m := y.ConfigTree.GetBaseValue(key)
	if m == "" {
		return defaultValue
	}
	return m
}

// FrameLoadInstanceHandler 自定义实例加载模式
type FrameLoadInstanceHandler interface {

	// LoadInstance 返回bool 自定义加载返回true 交给框架默认处理返回false
	LoadInstance(local *context.LocalStack, target *DynamicProxyInstanceNode,
		application *FrameApplication,
		applicationContext *FrameApplicationContext) bool

	Order() int
}

type DynamicProxyLinkedArray struct {
	FirstElement *DynamicProxyInstanceNode

	ElementMap map[string]*DynamicProxyInstanceNode

	LastElement *DynamicProxyInstanceNode
}

func (d *DynamicProxyLinkedArray) AddHead(node *DynamicProxyInstanceNode) {

	if node.Target != nil {
		target := node.Target
		node.rt = reflect.TypeOf(target)
		node.rv = reflect.ValueOf(target)
		fieldNum := node.rt.Elem().NumField()
		if fieldNum > 0 {
			for i := 0; i < fieldNum; i++ {
				field := node.rt.Elem().Field(i)
				if _, ok := field.Tag.Lookup(AutowiredInjectKey); ok {
					node.autowiredInjectField = append(node.autowiredInjectField, &field)
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

	if d.ElementMap == nil {
		d.ElementMap = make(map[string]*DynamicProxyInstanceNode)
	}

	d.ElementMap[node.Id] = node
}

func (d *DynamicProxyLinkedArray) Push(node *DynamicProxyInstanceNode) {

	if node.Target != nil {
		target := node.Target
		node.rt = reflect.TypeOf(target)
		node.rv = reflect.ValueOf(target)
		fieldNum := node.rt.Elem().NumField()
		if fieldNum > 0 {
			for i := 0; i < fieldNum; i++ {
				field := node.rt.Elem().Field(i)
				if _, ok := field.Tag.Lookup(AutowiredInjectKey); ok {
					node.autowiredInjectField = append(node.autowiredInjectField, &field)
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
	if d.ElementMap == nil {
		d.ElementMap = make(map[string]*DynamicProxyInstanceNode)
	}

	d.ElementMap[node.Id] = node

	if d.LastElement != nil {
		d.LastElement.Next = node
	}
	d.LastElement = node
}

type DynamicProxyInstanceNode struct {
	Target proxyclass.ProxyTarger

	Next *DynamicProxyInstanceNode

	Id string

	// Target 类型 push的时候设置  里面的值都是指针
	rt reflect.Type

	// Target 反射值 push的时候设置 里面的值都是指针
	rv reflect.Value

	// push的时候设置
	configInjectField []*reflect.StructField

	// push的时候设置
	autowiredInjectField []*reflect.StructField
}

type InsValueInjectTree struct {
	Root        *InsValueInjectTreeNode
	RefNode     map[string]*InsValueInjectTreeNode
	Environment *ConfigurableEnvironment
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
				name := util.ClassUtil.GetClassNameByType(field.Type.Elem())
				current.ObjectValue[name] = objectVal
			case reflect.Struct:
				name := util.ClassUtil.GetClassNameByType(field.Type)
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
		name := util.ClassUtil.GetClassNameByType(t.Elem())

		if v, ok1 := node.ObjectValue[name]; ok1 {
			m := reflect.ValueOf(v)
			return &m
		} else {
			return nil
		}
	case reflect.Struct:
		name := util.ClassUtil.GetClassNameByType(t)
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
				i.Environment.GetObjectValue(configkey, valmaprv.Interface())
				objectVal = valmaprv.Interface()
				val = &valmaprv
			case reflect.Ptr:
				v := reflect.New(field.Type.Elem())
				i.Environment.GetObjectValue(configkey, v.Interface())
				objectVal = v.Interface()
				val = &v
			case reflect.Struct:
				v := reflect.New(field.Type)
				i.Environment.GetObjectValue(configkey, v.Interface())
				objectVal = v.Interface()
				v1 := v.Elem()
				val = &v1
			default:
				v1 := i.Environment.GetBaseValue(configkey, defaultVal)
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
