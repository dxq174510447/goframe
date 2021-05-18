package dbcore

import (
	"context"
	"database/sql"
	"fmt"
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

type DatabaseFactory struct {
	dbUser string
	dbPwd  string
	dbName string
	dbPort string
	dbHost string
}

func (c *DatabaseFactory) NewDatabase() *sql.DB {
	//user:password@tcp(localhost:5555)/dbname?characterEncoding=UTF-8
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		c.dbUser, c.dbPwd, c.dbHost, c.dbPort, c.dbName,
	)
	fmt.Println(url)
	db, _ := sql.Open("mysql", url)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(20)
	return db
}

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
func OpenSqlConnection(readOnly int) *DatabaseConnection {

	ctx := context.Background()
	txOpt := sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  readOnly == 1,
	}

	// 取默认的key
	var key string = DataBaseDefaultKey
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

func init() {

	// 初始化默认数据源
	//var defaultFactory DatabaseFactory = DatabaseFactory{
	//	dbUser: util.ConfigUtil.Get("DB_USER", "platform"),
	//	dbPwd:  util.ConfigUtil.Get("DB_PASSWORD", "xxcxcx"),
	//	dbName: util.ConfigUtil.Get("DB_NAME", "plat_base1"),
	//	dbPort: util.ConfigUtil.Get("DB_PORT", "3306"),
	//	dbHost: util.ConfigUtil.Get("DB_HOST", "rm-bp1thh63s5tx33q0kio.mysql.rds.aliyuncs.com"),
	//}
	//db := defaultFactory.NewDatabase()
	//AddDatabaseRouter(DataBaseDefaultKey, db)

}
