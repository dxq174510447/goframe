package filter

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
)

// TxReadOnlyProxyFilter 只读事物
type TxReadOnlyProxyFilter struct {
	Logger logclass.AppLoger `FrameAutowired:""`
}

func (d *TxReadOnlyProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {

	var key string = "线程DB事物检查[PROPAGATION_READONLY]"

	if d.Logger.IsDebugEnable() {
		d.Logger.Debug(context, "%s begin", key)
		defer d.Logger.Debug(context, "%s end", key)
	}

	dbcon := dbcore.GetDbConnection(context)

	addNewConnection := false

	if dbcon == nil {
		d.Logger.Debug(context, "%s 当前线程未检查到dbconn", key)
		addNewConnection = true
	} else if dbcon != nil && !dbcon.TxOpt.ReadOnly {
		d.Logger.Debug(context, "%s 当前线程检测到dbconn connectid %s 但是是可写的 需要重新获取连接", key, dbcon.ConnectId)
		addNewConnection = true
	}

	var returnError interface{}
	var errorType int = 1 // 错误类型 1返回错误 2panic错误
	if addNewConnection {
		// 开启新的连接
		con := dbcore.OpenSqlConnection(context, 1)
		d.Logger.Debug(context, "%s 当前线程初始化新的 connectionId %s 并开启事物", key, con.ConnectId)

		// 将当前新的连接放入新的local变量中
		context.Push()
		dbcore.SetDbConnection(context, con) //连接不用释放 close方法没用

		// 启动事物
		con.BeginTransaction()
		defer func() {
			if err := recover(); err != nil {
				errorType = 2
				returnError = err
			}
			if returnError != nil {
				d.Logger.Debug(context, "%s 当前线程连接 connectionId %s 发生错误类型%d 回滚", key, con.ConnectId, errorType)
				con.Rollback()
				con.Close()
				context.Pop()
				if errorType == 2 {
					panic(returnError)
				}
			} else {
				d.Logger.Debug(context, "%s 当前线程连接 connectionId %s 提交事物", key, con.ConnectId)
				con.Commit()
				con.Close()
				context.Pop()
			}
		}()
	} else {
		// 当前线程已经绑定了连接 但是如果事物没有开启到话 就开启事物
		// 如果已经开启事物了的话 就不理会 交给开启事物的逻辑去处理
		// 连接也不管  交给开启连接的逻辑去处理
		if !dbcon.Transaction {
			d.Logger.Debug(context, "%s 当前线程检测到dbconn connectid %s 开启事物", key, dbcon.ConnectId)
			dbcon.BeginTransaction()
			defer func() {
				// 如果失败 回滚 继续往上抛错
				if err := recover(); err != nil {
					errorType = 2
					returnError = err
				}
				if returnError != nil {
					d.Logger.Debug(context, "%s 当前线程连接 connectionId %s 发生错误类型%d 回滚", key, dbcon.ConnectId, errorType)
					dbcon.Rollback()
					if errorType == 2 {
						panic(returnError)
					}
				} else {
					d.Logger.Debug(context, "%s 当前线程连接 connectionId %s 提交事物", key, dbcon.ConnectId)
					dbcon.Commit()
				}
			}()
		} else {
			d.Logger.Debug(context, "%s 当前线程检测到dbconn connectid %s 并且已经开启事物", key, dbcon.ConnectId)
		}
	}

	result := next.Execute(context, classInfo, methodInfo, invoker, arg)
	resultError := util.ClassUtil.GetErrorValueFromResult(result)
	if resultError != nil {
		returnError = resultError.Interface()
	}

	return result
}

func (d *TxReadOnlyProxyFilter) Order() int {
	return 3
}

func (d *TxReadOnlyProxyFilter) AnnotationMatch() []string {
	return []string{dbcore.TransactionReadOnly}
}

func (d *TxReadOnlyProxyFilter) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var txReadOnlyProxyFilter TxReadOnlyProxyFilter = TxReadOnlyProxyFilter{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&txReadOnlyProxyFilter))
}
