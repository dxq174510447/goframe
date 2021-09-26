package util

import (
	"fmt"
	"reflect"
)

var FrameErrorType reflect.Type = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()).Type()

type classUtil struct {
}

func (c *classUtil) GetErrorValueFromResult(result []reflect.Value) *reflect.Value {
	if len(result) == 0 {
		return nil
	}
	var returnError *reflect.Value
	// 从后往前
	i := len(result) - 1
	for ; i >= 0; i-- {
		row := result[i]
		if row.IsZero() {
			continue
		}
		if row.Type().Implements(FrameErrorType) {
			returnError = &row
			break
		}
	}
	return returnError
}

//GetClassNameV1 用来获取struct的全路径 传递指针
func (c *classUtil) GetClassNameV1(target interface{}) string {
	return c.GetClassNameByTypeV1(reflect.TypeOf(target))
}

//GetClassNameByTypeV1 接口，非指针到reflect.type类型
func (c *classUtil) GetClassNameByTypeV1(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return fmt.Sprintf("%s/%s", t.Elem().PkgPath(), t.Elem().Name())
	}
	return fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
}

//GetSimpleClassNameV1 用来获取struct的全路径 传递指针
func (c *classUtil) GetSimpleClassNameV1(target interface{}) string {
	return c.GetSimpleClassNameByTypeV1(reflect.TypeOf(target))
}

//GetSimpleClassNameByTypeV1 接口，非指针到reflect.type类型
func (c *classUtil) GetSimpleClassNameByTypeV1(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

func (c *classUtil) GetBeanNameV1(target interface{}) string {
	return c.GetBeanNameByTypeV1(reflect.TypeOf(target))
}

func (c *classUtil) GetBeanNameByTypeV1(t reflect.Type) string {
	m := c.GetSimpleClassNameByTypeV1(t)
	return StringUtil.FirstLower(m)
}

func (c *classUtil) GetFileNameV1(target interface{}) string {
	return c.GetFileNameByTypeV1(reflect.TypeOf(target))
}

func (c *classUtil) GetFileNameByTypeV1(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return fmt.Sprintf("%s/%s", t.Elem().PkgPath(), t.Elem().Name())
	}
	return fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
}

//func (c *classUtil) GetJavaClassNameByType(t reflect.Type) string {
//	m := t
//	if t.Kind() == reflect.Ptr {
//		m = t.Elem()
//	}
//	name := fmt.Sprintf("%s/%s", m.PkgPath(), m.Name())
//
//	name = strings.ReplaceAll(name, "/", ".")
//	p := strings.Index(name, ".main.golang.")
//	if p == -1 {
//		// 引用到时候
//		p = strings.Index(name, ".goframe.lib.")
//		if p == -1 {
//			p1 := strings.Index(name, ".")
//			if p1 == -1 {
//				return name
//			} else {
//				return name[p1+1:]
//			}
//		} else {
//			return name[p+1:]
//		}
//	} else {
//		return name[p+13:]
//	}
//}
//
//func (c *classUtil) GetJavaFileNameByType(name string) string {
//	p := strings.Index(name, "/main/golang/")
//	if p == -1 {
//		p = strings.Index(name, "/goframe/lib")
//		if p == -1 {
//			p1 := strings.Index(name, "/")
//			if p1 == -1 {
//				return name
//			} else {
//				return name[p1+1:]
//			}
//		} else {
//			return name[p+1:]
//		}
//	} else {
//		return name[p+13:]
//	}
//}

func (c *classUtil) IsNil(target interface{}) bool {
	if target == nil {
		return true
	}
	if reflect.ValueOf(target).IsNil() {
		return true
	}
	return false
}

var ClassUtil classUtil = classUtil{}
