package dbcore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	context2 "github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strconv"
	"time"
)

// dbRouter 关键字 对应的db源
var databaseRouter = make(map[string]*sql.DB)

func AddDatabaseRouter(key string, db *sql.DB) {
	databaseRouter[key] = db
}

//type DatabaseFactory struct {
//	DbUser string
//	DbPwd  string
//	DbName string
//	DbPort string
//	DbHost string
//	Prop   map[string]string
//}
//

type DatabaseConnection struct {
	Db    *sql.DB
	Con   *sql.Conn
	Ctx   context.Context
	TxOpt *sql.TxOptions

	// Transaction 当前连接是否开启事物
	Transaction bool
	ConnectId   string
}

func (d *DatabaseConnection) Close() {
	//不用释放
}

func (d *DatabaseConnection) BeginTransaction() {
	// 如果是只读事物 一定要try catch 最终要设置连接为 可读可写
	//TODO read only tx
	//SET autocommit = 0
	//tx,err := d.Con.BeginTx(d.Ctx,d.TxOpt) //好像版本有问题 触发失败
	//if err != nil {
	//	panic(err)
	//}
	//d.Transaction = tx
	if d.TxOpt.ReadOnly {
		d.Con.ExecContext(d.Ctx, "set session transaction read only")
	}

	d.Con.ExecContext(d.Ctx, "begin")
	d.Transaction = true
}

func (d *DatabaseConnection) Commit() {
	//d.Transaction.Commit()
	d.Con.ExecContext(d.Ctx, "commit")

	if d.TxOpt.ReadOnly {
		d.Con.ExecContext(d.Ctx, "set session transaction read write")
	}

	d.Transaction = false
}

func (d *DatabaseConnection) Rollback() {
	//d.Transaction.Rollback()
	d.Con.ExecContext(d.Ctx, "rollback")

	if d.TxOpt.ReadOnly {
		d.Con.ExecContext(d.Ctx, "set session transaction read write")
	}

	d.Transaction = false
}

// OpenSqlConnection 是否只读 1是 0否
func OpenSqlConnection(local *context2.LocalStack, readOnly int) *DatabaseConnection {

	ctx := context.Background()
	txOpt := sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  readOnly == 1,
	}

	// 取默认的key
	var key string = GetDbRouteKey(local)
	var db *sql.DB = databaseRouter[key]
	conn, err1 := db.Conn(ctx)

	if err1 != nil {
		fmt.Println(reflect.ValueOf(err1).Elem().String())
		fmt.Println(err1)
		panic(err1)
	}

	//获取connectid
	stmt2, err2 := conn.PrepareContext(ctx, "select connection_id()")

	if err2 != nil {
		panic(err2)
	}
	defer func() {
		stmt2.Close()
	}()
	result2 := stmt2.QueryRow()

	var connectId int = 0
	result2.Scan(&connectId)

	return &DatabaseConnection{
		Db:        db,
		Con:       conn,
		Ctx:       ctx,
		TxOpt:     &txOpt,
		ConnectId: strconv.Itoa(connectId),
	}
}

func NewDatabase(c *DatabaseConfig, key string) *DatabaseInstance {
	//user:password@tcp(localhost:5555)/dbname?characterEncoding=UTF-8
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		c.DbUser, c.DbPwd, c.DbHost, c.DbPort, c.DbName,
	)
	fmt.Println(url)
	db, _ := sql.Open("mysql", url)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(20)

	id := fmt.Sprintf("DbConnect_%s", key)

	return &DatabaseInstance{
		Db:     db,
		Id:     id,
		Config: c,
	}
}

type DatabaseConfig struct {
	DbUser    string
	DbPwd     string
	DbName    string
	DbPort    string
	DbHost    string
	Prop      map[string]string
	Proparray []string
	// proxy 全局唯一
}

type DatabaseInstance struct {
	Db     *sql.DB
	Id     string
	Config *DatabaseConfig
}

func (d *DatabaseInstance) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

type DatabaseFactory struct {
}

func (d *DatabaseFactory) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

func (d *DatabaseFactory) ProxyGet(local *context2.LocalStack, application1 *application.FrameApplication, applicationContext *application.FrameApplicationContext) proxyclass.ProxyTarger {
	var setting map[string]*DatabaseConfig = make(map[string]*DatabaseConfig)
	applicationContext.Environment.GetObjectValue("platform.datasource.config", setting)

	s, _ := json.Marshal(setting)
	fmt.Println("dbsetting--->", string(s))

	for k, v := range setting {
		db := NewDatabase(v, k)
		AddDatabaseRouter(k, db.Db)

		application1.FrameResource.ProxyInsPool.Push(&application.DynamicProxyInstanceNode{
			Target: db,
			Id:     db.Id,
		})
	}
	return nil
}

var databaseFactory DatabaseFactory = DatabaseFactory{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&databaseFactory))
}
