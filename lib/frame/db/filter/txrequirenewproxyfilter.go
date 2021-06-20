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

type TxRequireNewProxyFilter struct {
	Logger logclass.AppLoger `FrameAutowired:""`
}

func (d *TxRequireNewProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {

	var key string = "线程DB事物检查[PROPAGATION_REQUIRES_NEW]"

	if d.Logger.IsDebugEnable() {
		d.Logger.Debug(context, "%s begin", key)
		defer d.Logger.Debug(context, "%s end", key)
	}

	var returnError interface{}
	var errorType int = 1 // 错误类型 1返回错误 2panic错误

	con := dbcore.OpenSqlConnection(context, 0)
	d.Logger.Debug(context, "%s 当前线程初始化新的 connectionId %s 并开启事物", key, con.ConnectId)

	context.Push()
	dbcore.SetDbConnection(context, con) //连接不用释放 close方法没用

	// 启动事物
	con.BeginTransaction()

	defer func() {
		if err := recover(); err != nil {
			// 如果失败 回滚 继续往上抛错
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

	result := next.Execute(context, classInfo, methodInfo, invoker, arg)
	resultError := util.ClassUtil.GetErrorValueFromResult(result)
	if resultError != nil {
		returnError = resultError.Interface()
	}
	return result
}

func (d *TxRequireNewProxyFilter) Order() int {
	return 4
}

func (d *TxRequireNewProxyFilter) AnnotationMatch() []string {
	return []string{dbcore.TransactionRequireNew}
}

func (d *TxRequireNewProxyFilter) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var txRequireNewProxyFilter TxRequireNewProxyFilter = TxRequireNewProxyFilter{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&txRequireNewProxyFilter))
}
