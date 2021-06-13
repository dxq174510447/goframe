package filter

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"reflect"
)

type TxRequireProxyFilter struct {
	Logger application.AppLoger `FrameAutowired:""`
}

func (d *TxRequireProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {

	if d.Logger.IsTraceEnable() {
		d.Logger.Trace(context, "%s", "TxRequireProxyFilter begin")
		defer d.Logger.Trace(context, "%s", "TxRequireProxyFilter end")
	}

	dbcon := dbcore.GetDbConnection(context)

	addNewConnection := false

	if dbcon == nil {
		d.Logger.Trace(context, "%s", "当前线程未检测到dbconn")
		addNewConnection = true
	} else if dbcon != nil && dbcon.TxOpt.ReadOnly {
		d.Logger.Debug(context, "当前线程检测到dbconn connectid %s 但是是只读的 需要重新获取连接", dbcon.ConnectId)
		addNewConnection = true
	}

	if addNewConnection {
		// 开启新的连接
		con := dbcore.OpenSqlConnection(context, 0)
		d.Logger.Debug(context, "当前线程初始化新的 connectionId %s", con.ConnectId)

		// 将当前新的连接放入新的local变量中
		context.Push()
		dbcore.SetDbConnection(context, con) //连接不用释放 close方法没用

		// 启动事物
		con.BeginTransaction()
		defer func() {
			if err := recover(); err != nil {
				// 如果失败 回滚 继续往上抛错
				con.Rollback()

				con.Close()
				// 去除新的local变量
				context.Pop()

				panic(err)
			} else {
				// 没有错误 就提交
				con.Commit()

				con.Close()
				// 去除新的local变量
				context.Pop()
			}
		}()
	} else {
		// 如果当前已经绑定的连接 没有开启事物 就开启
		if !dbcon.Transaction {
			dbcon.BeginTransaction()
			defer func() {
				// 如果失败 回滚 继续往上抛错
				if err := recover(); err != nil {
					dbcon.Rollback()
					panic(err)
				} else {
					// 没有错误 就提交
					dbcon.Commit()
				}
			}()
		}
	}

	return next.Execute(context, classInfo, methodInfo, invoker, arg)
}

func (d *TxRequireProxyFilter) Order() int {
	return 5
}

func (d *TxRequireProxyFilter) AnnotationMatch() []string {
	return []string{dbcore.TransactionRequire}
}

func (d *TxRequireProxyFilter) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var txRequireProxyFilter TxRequireProxyFilter = TxRequireProxyFilter{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&txRequireProxyFilter))
}
