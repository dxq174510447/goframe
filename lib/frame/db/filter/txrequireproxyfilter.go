package filter

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/db/dbcore"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"reflect"
)

type TxRequireProxyFilter struct {
	Next proxy.ProxyFilter
}

func (d *TxRequireProxyFilter) Execute(context *context.LocalStack,
	classInfo *proxy.ProxyClass,
	methodInfo *proxy.ProxyMethod,
	invoker *reflect.Value,
	arg []reflect.Value) []reflect.Value {

	fmt.Printf("TxRequireProxyFilter begin \n")
	defer fmt.Printf("TxRequireProxyFilter end \n")

	dbcon := dbcore.GetDbConnection(context)

	addNewConnection := false

	if dbcon == nil {
		fmt.Printf("当前线程未检测到dbconn  \n")
		addNewConnection = true
	} else if dbcon != nil && dbcon.TxOpt.ReadOnly {
		fmt.Printf("当前线程检测到dbconn connectid %s 但是是只读的 需要重新获取连接 \n", dbcon.ConnectId)
		addNewConnection = true
	}

	if addNewConnection {
		// 开启新的连接
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

	return d.Next.Execute(context, classInfo, methodInfo, invoker, arg)
}

func (d *TxRequireProxyFilter) SetNext(next proxy.ProxyFilter) {
	d.Next = next
}

func (d *TxRequireProxyFilter) Order() int {
	return 5
}

var txRequireProxyFilter TxRequireProxyFilter = TxRequireProxyFilter{}

type TxRequireProxyFilterFactory struct {
}

func (d *TxRequireProxyFilterFactory) GetInstance(m map[string]interface{}) proxy.ProxyFilter {
	return proxy.ProxyFilter(&txRequireProxyFilter)
}

func (d *TxRequireProxyFilterFactory) AnnotationMatch() []string {
	return []string{dbcore.TransactionRequire}
}

var txRequireProxyFilterFactory TxRequireProxyFilterFactory = TxRequireProxyFilterFactory{}

func init() {
	proxy.AddProxyFilterFactory(proxy.ProxyFilterFactory(&txRequireProxyFilterFactory))
}
