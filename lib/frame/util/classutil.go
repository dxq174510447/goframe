package util

import (
	"fmt"
	"reflect"
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

var ClassUtil classUtil = classUtil{}
