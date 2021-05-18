package filter

import (
	"fmt"
	"goframe/lib/frame/context"
	"goframe/lib/frame/db/dbcore"
	"goframe/lib/frame/proxy"
	"reflect"
)

type DaoConnectProxyFilter struct {
	Next proxy.ProxyFilter
}

func (d *DaoConnectProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxy.ProxyClass,
	methodInfo *proxy.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value) []reflect.Value {

	fmt.Printf("DaoConnectProxyFilter begin \n")
	defer fmt.Printf("DaoConnectProxyFilter end \n")

	dbcon := dbcore.GetDbConnection(context)

	if dbcon != nil {
		fmt.Printf("当前线程检测到dbconn connectid %s \n", dbcon.ConnectId)
	}

	if dbcon == nil {
		con := dbcore.OpenSqlConnection(0)
		fmt.Printf("当前线程未检测到dbconn 初始化之后 connectid %s \n", con.ConnectId)

		context.Push()
		dbcore.SetDbConnection(context, con) //连接不用释放 close方法没用

		defer func() {
			con.Close()
			context.Pop()
		}()
	}

	return d.Next.Execute(context, classInfo, methodInfo, invoker, arg)
}

func (d *DaoConnectProxyFilter) SetNext(next proxy.ProxyFilter) {
	d.Next = next
}

func (d *DaoConnectProxyFilter) Order() int {
	return 15
}

var daoConnectProxyFilter DaoConnectProxyFilter = DaoConnectProxyFilter{}

type DaoConnectProxyFilterFactory struct {
}

func (d *DaoConnectProxyFilterFactory) GetInstance(m map[string]interface{}) proxy.ProxyFilter {
	return proxy.ProxyFilter(&daoConnectProxyFilter)
}

func (d *DaoConnectProxyFilterFactory) AnnotationMatch() []string {
	return []string{proxy.AnnotationDao}
}

var daoConnectProxyFilterFactory DaoConnectProxyFilterFactory = DaoConnectProxyFilterFactory{}

func init() {
	proxy.AddProxyFilterFactory(proxy.ProxyFilterFactory(&daoConnectProxyFilterFactory))
}
