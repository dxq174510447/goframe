package util

import (
	"fmt"
	"reflect"
	"strings"
)

var FrameErrorType reflect.Type = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()).Type()

type classUtil struct {
}

//GetClassName 用来获取struct的全路径 传递指针
func (c *classUtil) GetClassName(target interface{}) string {
	t := reflect.ValueOf(target).Elem().Type()
	return fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
}

func (c *classUtil) GetClassNameByType(t reflect.Type) string {
	return fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
}

func (c *classUtil) GetJavaClassNameByType(t reflect.Type) string {
	m := t
	if t.Kind() == reflect.Ptr {
		m = t.Elem()
	}
	name := fmt.Sprintf("%s/%s", m.PkgPath(), m.Name())

	name = strings.ReplaceAll(name, "/", ".")
	p := strings.Index(name, ".main.golang.")
	if p == -1 {
		// 引用到时候
		p = strings.Index(name, ".goframe.lib.")
		if p == -1 {
			p1 := strings.Index(name, ".")
			if p1 == -1 {
				return name
			} else {
				return name[p1+1:]
			}
		} else {
			return name[p+1:]
		}
	} else {
		return name[p+13:]
	}
}

func (c *classUtil) GetJavaFileNameByType(name string) string {
	p := strings.Index(name, "/main/golang/")
	if p == -1 {
		p = strings.Index(name, "/goframe/lib")
		if p == -1 {
			p1 := strings.Index(name, "/")
			if p1 == -1 {
				return name
			} else {
				return name[p1+1:]
			}
		} else {
			return name[p+1:]
		}
	} else {
		return name[p+13:]
	}
}

var ClassUtil classUtil = classUtil{}
