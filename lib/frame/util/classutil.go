package util

import (
	"fmt"
	"reflect"
	"strings"
)

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
	name := fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
	name = strings.ReplaceAll(name, "/", ".")
	p := strings.Index(name, "main.golang")
	if p == -1 {
		p1 := strings.Index(name, ".")
		if p1 == -1 {
			return name
		} else {
			return name[p1+1:]
		}
	} else {
		return name[p+12:]
	}
}

var ClassUtil classUtil = classUtil{}
