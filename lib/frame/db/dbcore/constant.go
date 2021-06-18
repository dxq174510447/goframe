package dbcore

import (
	"database/sql"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"reflect"
	"time"
)

const (
	// TransactionReadOnly 只读 用于注解
	TransactionReadOnly = "TransactionReadOnly_"

	// TransactionRequire 如果有事物的话无需创建 没有的话就创建一个
	TransactionRequire = "TransactionRequire_"

	// TransactionRequireNew 如果有事物的话 在创建一个
	TransactionRequireNew = "TransactionRequireNew_"

	//DataBaseDefaultKey db 路由 默认key
	DataBaseDefaultKey = "default"

	// DataBaseConnectKey 上下文中保存的数据库连接
	DataBaseConnectKey = "DataBaseConnectKey_"

	// MapperErrorHandlerFlagKey 上下文中保存mapper错误处理方式 0panic 1 return error
	MapperErrorHandlerFlagKey = "MapperErrorHandlerFlagKey_"
	DataBaseTxKey             = "DataBaseTxKey_"
	DataBaseRouteKey          = "DataBaseRouteKey_"

	DbConfigMaxOpenConnsKey = "DbConfigMaxOpenConnsKey_"
	DbConfigMaxIdleConnsKey = "DbConfigMaxIdleConnsKey_"
)

const (
	SqlTypeInsert = "insert_"
	SqlTypeUpdate = "update_"
	SqlTypeDelete = "delete_"
	SqlTypeSelect = "select_"
	SqlTypeSql    = "sql_"
	// dao 中参数配置
	AnnotationSqlProviderConfig         = "AnnotationSqlProviderConfig_"
	AnnotationSqlProviderConfigValueKey = "AnnotationSqlProviderConfigValueKey_"

	AnnotationDaoConfigValueKey = "AnnotationDaoConfigValueKey_"
)

var SqlNullInt32Type reflect.Type = reflect.TypeOf(sql.NullInt32{})
var SqlNullTimeType reflect.Type = reflect.TypeOf(sql.NullTime{})
var SqlNullFloat64Type reflect.Type = reflect.TypeOf(sql.NullFloat64{})
var SqlNullStringType reflect.Type = reflect.TypeOf(sql.NullString{})
var SqlNullInt64Type reflect.Type = reflect.TypeOf(sql.NullInt64{})
var SqlNullBoolType reflect.Type = reflect.TypeOf(sql.NullBool{})
var GoTimeType reflect.Type = reflect.TypeOf(time.Time{})

func SetDbConnection(local *context.LocalStack, con *DatabaseConnection) {
	local.Set(DataBaseConnectKey, con)
}

func GetDbConnection(local *context.LocalStack) *DatabaseConnection {
	db := local.Get(DataBaseConnectKey)
	if db == nil {
		return nil
	}
	return db.(*DatabaseConnection)
}

// GetErrorHandleFlag flag 0 标记panic 1标记 return error
func GetErrorHandleFlag(local *context.LocalStack) int {
	flag := local.Get(MapperErrorHandlerFlagKey)
	if flag == nil {
		return 0
	}
	return flag.(int)
}

// SetErrorHandleFlag flag 0 标记panic 1标记 return error
func SetErrorHandleFlag(local *context.LocalStack, flag int) {
	local.Set(MapperErrorHandlerFlagKey, flag)
}

func SetDbRouteKey(local *context.LocalStack, key string) {
	local.Set(DataBaseRouteKey, key)
}

func GetDbRouteKey(local *context.LocalStack) string {
	m := local.Get(DataBaseRouteKey)
	if m == nil {
		return DataBaseDefaultKey
	}
	return m.(string)
}
