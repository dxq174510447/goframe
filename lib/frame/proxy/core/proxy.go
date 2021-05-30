package core

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type methodInvoke struct {
	target   interface{}
	clazz    *proxyclass.ProxyClass
	method   *proxyclass.ProxyMethod
	invoker  *reflect.Value
	filters  []*proxyclass.ProxyFilterWrapper
	initLock sync.Mutex
}

func (m *methodInvoke) initFilter() {
	if m.filters != nil && len(m.filters) > 0 {
		return
	}
	m.initLock.Lock()
	defer m.initLock.Unlock()

	if m.filters != nil && len(m.filters) > 0 {
		return
	}

	fmt.Println("初始化")

	var fs []proxyclass.ProxyFilter
	var hasAdd map[string]string = make(map[string]string)

	// method level
	if m.method != nil && len(m.method.Annotations) > 0 {
		for _, annotation := range m.method.Annotations {
			if factorys, ok := methodFilter[annotation.Name]; ok {
				for _, factory := range factorys {
					fs = append(fs, factory)
					hasAdd[annotation.Name] = "1"
				}
			}
		}
	}

	if m.clazz != nil && len(m.clazz.Annotations) > 0 {
		for _, annotation := range m.clazz.Annotations {
			if _, ok := hasAdd[annotation.Name]; ok {
				continue
			}
			if factorys, ok := methodFilter[annotation.Name]; ok {
				for _, factory := range factorys {
					fs = append(fs, factory)
				}
			}
		}
	}

	if len(globalFilter) > 0 {
		fs = append(fs, globalFilter...)
	}

	if len(fs) > 1 {
		sort.Slice(fs, func(i, j int) bool {
			return fs[i].Order() < fs[j].Order()
		})
	}

	fmt.Println(len(fs))
	wrapfs := make([]*proxyclass.ProxyFilterWrapper, len(fs), len(fs))
	for i := 0; i < len(fs); i++ {
		wrapfs[i] = &proxyclass.ProxyFilterWrapper{
			Target: fs[i],
		}
	}

	l := len(wrapfs)
	for i, f := range wrapfs {
		if i == (l - 1) {
			break
		}
		f.Next = wrapfs[i+1]
	}

	m.filters = wrapfs
}
func (m *methodInvoke) invoke(context *context.LocalStack, args []reflect.Value) []reflect.Value {
	m.initFilter()
	//fmt.Println(len(m.filters))
	for _, r := range m.filters {
		fmt.Println(reflect.ValueOf(r).Elem().Type().Name())
	}
	return m.filters[0].Execute(context,
		m.clazz,
		m.method,
		m.invoker,
		args,
	)
}

func newMethodInvoke(
	target interface{},
	clazz *proxyclass.ProxyClass,
	method *proxyclass.ProxyMethod,
	invoker *reflect.Value) *methodInvoke {
	return &methodInvoke{
		target:  target,
		clazz:   clazz,
		method:  method,
		invoker: invoker,
	}
}

// classProxy 好像没地方用到 key是全路径 GetClassName
var classProxy map[string]*proxyclass.ProxyClass = make(map[string]*proxyclass.ProxyClass)

var methodFilter map[string][]proxyclass.ProxyFilter = make(map[string][]proxyclass.ProxyFilter)

var globalFilter []proxyclass.ProxyFilter = make([]proxyclass.ProxyFilter, 0, 0)

// AddAopProxyFilter 添加拦截器
func AddAopProxyFilter(target proxyclass.ProxyFilter) {
	AddClassProxy(target.(proxyclass.ProxyTarger))

	match := target.AnnotationMatch()
	if len(match) == 0 {
		globalFilter = append(globalFilter, target)
		return
	}

	for _, annotation := range match {
		if v, ok := methodFilter[annotation]; ok {
			methodFilter[annotation] = append(v, target)
		} else {
			methodFilter[annotation] = []proxyclass.ProxyFilter{target}
		}
	}

}

func AddClassProxy(target proxyclass.ProxyTarger) {
	clazz := target.ProxyTarget()
	clazzName := util.ClassUtil.GetClassName(target)

	//添加到映射中
	if clazz == nil {
		clazz = &proxyclass.ProxyClass{}
	}
	clazz.Target = target
	clazz.Name = clazzName
	classProxy[clazzName] = clazz

	//获取对象方法设置
	methodRef := make(map[string]*proxyclass.ProxyMethod)
	if len(clazz.Methods) != 0 {
		for _, md := range clazz.Methods {
			methodRef[md.Name] = md
		}
	}

	//解析字段方法 包裹一层
	rv := reflect.ValueOf(target)
	rt := rv.Elem().Type()
	if m1 := rt.NumField(); m1 > 0 {
		for i := 0; i < m1; i++ {
			field := rt.Field(i)
			proxyStructFuncField(target, &rv, rt, methodRef, target, &rv, rt, methodRef, &field)
		}
	}
}

// 重新设置struct field func 设置代理链
func proxyStructFuncField(target proxyclass.ProxyTarger,
	targetValue *reflect.Value,
	targetType reflect.Type,
	targetMethod map[string]*proxyclass.ProxyMethod,
	currentTarget proxyclass.ProxyTarger,
	currentTargetValue *reflect.Value,
	currentTargetType reflect.Type,
	currentTargetMethod map[string]*proxyclass.ProxyMethod,
	field *reflect.StructField) {

	var isproxyfield bool = false
	if field.Type.Kind() == reflect.Func {
		isproxyfield = true
	} else if field.Type.Kind() == reflect.Struct {
		//fmt.Println(field.Type.Kind(),field.Name
		// currentTargetValue is ptr
		fieldValue := currentTargetValue.Elem().FieldByName(field.Name)
		fieldType, _ := currentTargetType.FieldByName(field.Name)
		if pt, ok := fieldValue.Addr().Interface().(proxyclass.ProxyTarger); ok {
			if m1 := fieldType.Type.NumField(); m1 > 0 {

				methodRef := make(map[string]*proxyclass.ProxyMethod)
				if len(pt.ProxyTarget().Methods) != 0 {
					for _, md := range pt.ProxyTarget().Methods {
						methodRef[md.Name] = md
					}
				}

				for i := 0; i < m1; i++ {
					hhfield := fieldType.Type.Field(i)
					ptrval := fieldValue.Addr()
					proxyStructFuncField(target, targetValue, targetType, targetMethod,
						pt, &ptrval, fieldType.Type, methodRef, &hhfield)
				}
			}
			return
		}
	}

	if !isproxyfield {
		return
	}

	call := currentTargetValue.Elem().FieldByName(field.Name)
	oldCall := reflect.ValueOf(call.Interface())

	methodName := strings.ReplaceAll(field.Name, "_", "")
	var methodSetting *proxyclass.ProxyMethod

	if methodSetting1, ok1 := currentTargetMethod[methodName]; !ok1 {
		if methodSetting2, ok2 := targetMethod[methodName]; !ok2 {
			methodSetting = methodSetting2
		}
	} else {
		methodSetting = methodSetting1
	}
	if methodSetting == nil {
		methodSetting = &proxyclass.ProxyMethod{Name: methodName}
	}

	invoker := newMethodInvoke(currentTarget, target.ProxyTarget(), methodSetting, &oldCall)

	proxyCall := func(command *methodInvoke) reflect.Value {
		newCall := reflect.MakeFunc(field.Type, func(in []reflect.Value) []reflect.Value {
			fmt.Printf(" %s %s agent begin \n", invoker.clazz.Name, invoker.method.Name)
			defer fmt.Printf(" %s %s agent end \n", invoker.clazz.Name, invoker.method.Name)
			return command.invoke(in[0].Interface().(*context.LocalStack), in)
		})
		return newCall
	}(invoker)
	call.Set(proxyCall)
}

func getTargetValue(target interface{}, name string) interface{} {
	v := reflect.ValueOf(target)
	switch v.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		return target
	case reflect.Map:
		m := target.(map[string]interface{})
		if strings.Index(name, "[") > 0 {
			p1 := strings.Index(name, "[")
			p2 := strings.Index(name, "]")
			field := name[0:p1]
			index := name[p1+1 : p2]
			index1, _ := strconv.Atoi(index)
			return reflect.ValueOf(m[field]).Index(index1).Interface()
		} else {
			return m[name]
		}
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct {
			if strings.Index(name, "[") > 0 {
				p1 := strings.Index(name, "[")
				p2 := strings.Index(name, "]")
				field := name[0:p1]
				index := name[p1+1 : p2]
				index1, _ := strconv.Atoi(index)
				return v.Elem().FieldByName(field).Index(index1).Interface()
			} else {
				return v.Elem().FieldByName(name).Interface()
			}
		}
	}
	panic(fmt.Sprintf("%s找不到对应属性", name))
}

// GetVariableValue target 可能map接口 基础类型 指针结构体类型
func GetVariableValue(target interface{}, name string) interface{} {

	keys := strings.Split(name, ".")

	l := len(keys)
	if l == 1 {
		return getTargetValue(target, name)
	} else {
		nt := target
		for i := 0; i < l; i++ {
			nt = getTargetValue(nt, keys[i])
			//中间值为nil 就panic
			if i < (l-1) && reflect.ValueOf(nt).IsZero() {
				panic(fmt.Sprintf("sql %s is nil value", name))
			}
		}
		return nt
	}

}

func GetStructField(rtType reflect.Type) map[string]reflect.StructField {
	ref := make(map[string]reflect.StructField)
	switch rtType.Kind() {
	case reflect.Slice:
		if rtType.Elem().Kind() == reflect.Struct {
			n := rtType.Elem().NumField()
			for i := 0; i < n; i++ {
				sf := rtType.Elem().Field(i)
				ref[sf.Name] = sf
			}
			return ref
		} else if rtType.Elem().Kind() == reflect.Ptr && rtType.Elem().Elem().Kind() == reflect.Struct {
			n := rtType.Elem().Elem().NumField()
			for i := 0; i < n; i++ {
				sf := rtType.Elem().Elem().Field(i)
				ref[sf.Name] = sf
			}
			return ref
		} else {
			return nil
		}
	case reflect.Ptr:
		if rtType.Elem().Kind() != reflect.Struct {
			return nil
		} else {
			n := rtType.Elem().NumField()
			for i := 0; i < n; i++ {
				sf := rtType.Elem().Field(i)
				ref[sf.Name] = sf
			}
			return ref
		}
	default:
		return nil
	}

}

// GetMethodReturnDefaultValue int:默认值 float64:默认值 string:默认值 slice:默认值 ptr(struct)空的指针 map:默认值
// 当有错误的时候 返回这个默认结果 和 错误
func GetMethodReturnDefaultValue(rtType reflect.Type) *reflect.Value {
	var result reflect.Value
	switch rtType.Kind() {
	case reflect.String:
		result = reflect.ValueOf("")
	case reflect.Int64:
		result = reflect.ValueOf(int64(0))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		result = reflect.ValueOf(0)
	case reflect.Float32, reflect.Float64:
		result = reflect.ValueOf(0.0)
	case reflect.Map:
		v := reflect.MakeMap(rtType)
		result = v
	case reflect.Slice:
		result = reflect.MakeSlice(rtType, 0, 0)
	case reflect.Ptr:
		result = reflect.New(rtType).Elem()
	case reflect.Struct:
		result = reflect.New(rtType).Elem()
	case reflect.Interface:
		//base Get
		return nil
	default:
		panic(fmt.Sprintf("%s找不到对应默认值", rtType.String()))
	}
	return &result
}

var matchAllCap = regexp.MustCompile(`[^A-Za-z0-9]+`)

func GetCamelCaseName(str string) string {
	st := matchAllCap.Split(str, -1)
	for k, s := range st {
		st[k] = strings.Title(strings.ToLower(s))
	}
	return strings.Join(st, "")
}
