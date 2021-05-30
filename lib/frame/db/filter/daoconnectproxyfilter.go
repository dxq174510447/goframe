package filter

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"reflect"
)

type DaoConnectProxyFilter struct {
}

func (d *DaoConnectProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {

	fmt.Printf("DaoConnectProxyFilter begin \n")
	defer fmt.Printf("DaoConnectProxyFilter end \n")

	dbcon := dbcore.GetDbConnection(context)

	if dbcon != nil {
		fmt.Printf("当前线程检测到dbconn connectid %s \n", dbcon.ConnectId)
	}

	if dbcon == nil {
		con := dbcore.OpenSqlConnection(context, 0)
		fmt.Printf("当前线程未检测到dbconn 初始化之后 connectid %s \n", con.ConnectId)

		context.Push()
		dbcore.SetDbConnection(context, con) //连接不用释放 close方法没用

		defer func() {
			con.Close()
			context.Pop()
		}()
	}

	return next.Execute(context, classInfo, methodInfo, invoker, arg)
}

func (d *DaoConnectProxyFilter) Order() int {
	return 15
}

func (d *DaoConnectProxyFilter) AnnotationMatch() []string {
	return []string{proxyclass.AnnotationDao}
}

func (d *DaoConnectProxyFilter) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var daoConnectProxyFilter DaoConnectProxyFilter = DaoConnectProxyFilter{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&daoConnectProxyFilter))
}
