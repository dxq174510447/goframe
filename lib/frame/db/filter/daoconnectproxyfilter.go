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

	var key string = "线程DB连接检查"
	if d.Logger.IsDebugEnable() {
		d.Logger.Debug(context, "%s begin", key)
		defer d.Logger.Debug(context, "%s end", key)
	}

	dbcon := dbcore.GetDbConnection(context)

	if dbcon != nil {
		d.Logger.Debug(context, "%s 当前线程检测到dbconn connectid %s", key, dbcon.ConnectId)
	} else {
		con := dbcore.OpenSqlConnection(context, 0)
		d.Logger.Debug(context, "%s 当前线程未检测到dbconn 新建的连接 connectid %s", key, con.ConnectId)

		context.Push()
		dbcore.SetDbConnection(context, con) //连接不用释放 close方法没用

		defer func() {
			// 在这里创建的连接 不管结果成功或者失败 以何种方式失败 都释放连接
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
