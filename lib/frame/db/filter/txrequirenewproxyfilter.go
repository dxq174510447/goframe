package filter

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"reflect"
)

type TxRequireNewProxyFilter struct {
	Next proxy.ProxyFilter
}

func (d *TxRequireNewProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxy.ProxyClass,
	methodInfo *proxy.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value) []reflect.Value {

	fmt.Printf("TxRequireNewProxyFilter begin \n")
	defer fmt.Printf("TxRequireNewProxyFilter end \n")

	// 无论线程变量中有没有连接都创建一个新的
	con := dbcore.OpenSqlConnection(0)
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

	return d.Next.Execute(context, classInfo, methodInfo, invoker, arg)

}

func (d *TxRequireNewProxyFilter) SetNext(next proxy.ProxyFilter) {
	d.Next = next
}

func (d *TxRequireNewProxyFilter) Order() int {
	return 4
}

var txRequireNewProxyFilter TxRequireNewProxyFilter = TxRequireNewProxyFilter{}

type TxRequireNewProxyFilterFactory struct {
}

func (d *TxRequireNewProxyFilterFactory) GetInstance(m map[string]interface{}) proxy.ProxyFilter {
	return proxy.ProxyFilter(&txRequireNewProxyFilter)
}

func (d *TxRequireNewProxyFilterFactory) AnnotationMatch() []string {
	return []string{dbcore.TransactionRequireNew}
}

var txRequireNewProxyFilterFactory TxRequireNewProxyFilterFactory = TxRequireNewProxyFilterFactory{}

func init() {
	proxy.AddProxyFilterFactory(proxy.ProxyFilterFactory(&txRequireNewProxyFilterFactory))
}
