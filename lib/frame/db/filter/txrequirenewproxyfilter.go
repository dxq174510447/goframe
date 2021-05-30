package filter

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"reflect"
)

type TxRequireNewProxyFilter struct {
}

func (d *TxRequireNewProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxyclass.ProxyClass,
	methodInfo *proxyclass.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value, next *proxyclass.ProxyFilterWrapper) []reflect.Value {

	fmt.Printf("TxRequireNewProxyFilter begin \n")
	defer fmt.Printf("TxRequireNewProxyFilter end \n")

	// 无论线程变量中有没有连接都创建一个新的
	con := dbcore.OpenSqlConnection(context, 0)
	fmt.Printf("当前线程初始化新的 connectionId %s \n", con.ConnectId)

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

	return next.Execute(context, classInfo, methodInfo, invoker, arg)

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
