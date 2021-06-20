package filter

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"reflect"
)

type DaoConnectProxyFilter struct {
	Logger logclass.AppLoger `FrameAutowired:""`
}

func (d *DaoConnectProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {

	if d.Logger.IsTraceEnable() {
		d.Logger.Trace(context, "%s", "DaoConnectProxyFilter begin")
		defer d.Logger.Trace(context, "%s", "DaoConnectProxyFilter end ")
	}

	dbcon := dbcore.GetDbConnection(context)

	if dbcon != nil {
		d.Logger.Trace(context, "当前线程检测到dbconn connectid %s", dbcon.ConnectId)
	}

	if dbcon == nil {
		con := dbcore.OpenSqlConnection(context, 0)
		d.Logger.Debug(context, "当前线程未检测到dbconn 初始化之后 connectid %s", con.ConnectId)

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
