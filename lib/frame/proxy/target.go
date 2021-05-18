package proxy

type AnnotationClass struct {
	Name  string
	Value map[string]interface{}
}

type ProxyClass struct {
	Name        string
	Target      interface{}
	Methods     []*ProxyMethod
	Annotations []*AnnotationClass
}

type ProxyMethod struct {
	Name        string
	Annotations []*AnnotationClass
}

// ProxyTarger 代理对象都必须继承实现的方法
type ProxyTarger interface {
	ProxyTarget() *ProxyClass
}
