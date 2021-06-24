package dbcore

import (
	"bytes"
	"database/sql"
	"encoding/xml"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/exception"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/core"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

//type DaoProxyTarger interface {
//	proxy.ProxyTarger
//	DaoXml() string
//	DaoEntity() interface{}
//}

type DaoConfig struct {
	// sql文件
	Xml string
	// entity 指针
	Entity interface{}
}

type SqlProviderConfig struct {
	Param string
}

type MapperElementXml struct {
	Id      string `xml:"id,attr"`
	Sql     string `xml:",innerxml"`
	SqlType string
	Tpl     *template.Template
}

type MapperXml struct {
	Sql []*MapperElementXml `xml:"sql"`

	SelectSql []*MapperElementXml `xml:"select"`

	UpdateSql []*MapperElementXml `xml:"update"`

	DeleteSql []*MapperElementXml `xml:"delete"`

	InsertSql []*MapperElementXml `xml:"insert"`
}

type MapperFactory struct {
	importTagRegexp *regexp.Regexp
	columnTagRegexp *regexp.Regexp
}

func (m *MapperFactory) ReplaceImportTag(sql string, refs map[string]*MapperElementXml, table *TableDef) string {
	ns := m.importTagRegexp.ReplaceAllStringFunc(sql, func(str string) string {
		str1 := m.importTagRegexp.FindStringSubmatch(str)
		if s, ok := refs[str1[1]]; ok {
			return s.Sql
		}
		return ""
	})

	if table != nil {
		ns = m.columnTagRegexp.ReplaceAllStringFunc(ns, func(str string) string {
			str1 := m.columnTagRegexp.FindStringSubmatch(str)
			var buffer []string
			if len(str1) == 0 || str1[0] == "" {
				for _, col := range table.Columns {
					buffer = append(buffer, fmt.Sprintf("%c%s%c", 96, col.FieldName, 96))
				}
			} else {
				for _, col := range table.Columns {
					buffer = append(buffer, fmt.Sprintf("%s.%c%s%c", str1[0], 96, col.FieldName, 96))
				}
			}
			return strings.Join(buffer, ",")
		})
	}

	reg := regexp.MustCompile(`(?m)(^\s+|\s+$)`)
	ns = reg.ReplaceAllString(ns, " ")
	ns = strings.ReplaceAll(ns, "<![CDATA[", "")
	ns = strings.ReplaceAll(ns, "]]>", "")
	return ns
}

func (m *MapperFactory) ParseXml(target proxyclass.ProxyTarger, content string, table *TableDef) map[string]*MapperElementXml {
	mapper := &MapperXml{}
	err := xml.Unmarshal([]byte(content), mapper)

	if err != nil {
		panic(err)
	}
	refs := make(map[string]*MapperElementXml)

	if len(mapper.Sql) > 0 {
		for _, ele := range mapper.Sql {
			ele.SqlType = SqlTypeSql
			refs[ele.Id] = ele
		}
	}

	if len(mapper.UpdateSql) > 0 {
		for _, ele := range mapper.UpdateSql {
			ele.SqlType = SqlTypeUpdate
			ele.Sql = m.ReplaceImportTag(ele.Sql, refs, table)
			ele.Tpl = template.Must(template.New(fmt.Sprintf("%s-%s", util.ClassUtil.GetClassName(target), ele.Id)).Parse(ele.Sql))
			refs[ele.Id] = ele
		}
	}

	if len(mapper.InsertSql) > 0 {
		for _, ele := range mapper.InsertSql {
			ele.SqlType = SqlTypeInsert
			ele.Sql = m.ReplaceImportTag(ele.Sql, refs, table)
			ele.Tpl = template.Must(template.New(fmt.Sprintf("%s-%s", util.ClassUtil.GetClassName(target), ele.Id)).Parse(ele.Sql))
			refs[ele.Id] = ele
		}
	}

	if len(mapper.SelectSql) > 0 {
		for _, ele := range mapper.SelectSql {
			ele.SqlType = SqlTypeSelect
			ele.Sql = m.ReplaceImportTag(ele.Sql, refs, table)
			ele.Tpl = template.Must(template.New(fmt.Sprintf("%s-%s", util.ClassUtil.GetClassName(target), ele.Id)).Parse(ele.Sql))
			refs[ele.Id] = ele
		}
	}

	if len(mapper.DeleteSql) > 0 {
		for _, ele := range mapper.DeleteSql {
			ele.SqlType = SqlTypeDelete
			ele.Sql = m.ReplaceImportTag(ele.Sql, refs, table)
			ele.Tpl = template.Must(template.New(fmt.Sprintf("%s-%s", util.ClassUtil.GetClassName(target), ele.Id)).Parse(ele.Sql))
			refs[ele.Id] = ele
		}
	}

	return refs
}

var mapperFactory MapperFactory = MapperFactory{
	importTagRegexp: regexp.MustCompile(`<include refid="(\S+)">\s*</include>`),
	columnTagRegexp: regexp.MustCompile(`<column alias="(\S+)">\s*</column>`),
}

func GetMapperFactory() *MapperFactory {
	return &mapperFactory
}

type sqlColumnType struct {
	column      *sql.ColumnType
	field       *reflect.StructField
	defaultType reflect.Type
}

type sqlInvoke struct {
	Logger         logclass.AppLoger
	target         interface{}
	clazz          *proxyclass.ProxyClass
	method         *proxyclass.ProxyMethod
	mapper         map[string]*MapperElementXml
	providerConfig *SqlProviderConfig
	//slice int string ptr float64
	returnSqlType reflect.Type

	//具体返回的类型 如果返回的是指针 就对应的是结构体 如果返回的是slice 就对应的里面的元素类型，如果元素是指针就是对应的结构体 否则就是int,string等
	returnSqlElementType reflect.Type
	defaultReturnValue   *reflect.Value
	//如果返回的是结构体类型 字段field
	structFieldMap map[string]reflect.StructField
	sqlFieldMap    []*sqlColumnType
	entityType     reflect.Type
	//base sql
	baseXmlEle map[string]*MapperElementXml
	//如果使用到entity entity列定义
	tableDef *TableDef
	//当前方法是不是base里面的method
	baseMethod bool
}

func (s *sqlInvoke) invoke(context *context.LocalStack, args []reflect.Value) []reflect.Value {
	methodName := s.method.Name
	if ele, ok := s.mapper[methodName]; ok {
		s.Logger.Debug(context, "%s 对应sql处理节点 %s", methodName, ele.Id)
		switch ele.SqlType {
		case SqlTypeSelect:
			return s.invokeSelect(context, args, ele)
		case SqlTypeUpdate:
			return s.invokeUpdate(context, args, ele)
		case SqlTypeInsert:
			return s.invokeInsert(context, args, ele)
		case SqlTypeDelete:
			return s.invokeDelete(context, args, ele)
		}
	} else {
		s.Logger.Error(context, nil, "%s 找不到对应sql处理节点", methodName)
	}
	return nil
}

func (s *sqlInvoke) getSqlFromTpl(context *context.LocalStack, args []reflect.Value, sql *MapperElementXml) (string, error) {

	// 去除局部变量参数
	if len(args) <= 1 {
		return sql.Sql, nil
	}

	// 只有一个参数 结构体 基础类型 string
	var params []string
	if s.providerConfig != nil && s.providerConfig.Param != "" {
		params = strings.Split(s.providerConfig.Param, ",")
	}
	if len(args) == 2 {
		if len(params) >= 2 && params[1] != "_" && params[1] != "" {
			root := make(map[string]interface{})
			root[params[1]] = args[1].Interface()
			buf := &bytes.Buffer{}
			err := sql.Tpl.Execute(buf, root)
			if err != nil {
				return "", err
			}
			return RemoveEmptyRow(buf.String()), nil
		} else {
			buf := &bytes.Buffer{}
			err := sql.Tpl.Execute(buf, args[1].Interface())
			if err != nil {
				return "", err
			}
			return RemoveEmptyRow(buf.String()), nil
		}
	}
	// > 2
	root := make(map[string]interface{})
	for i := 1; i < len(params); i++ {
		if params[i] != "_" && params[i] != "" {
			root[params[i]] = args[i].Interface()
		}
	}
	buf := &bytes.Buffer{}
	err := sql.Tpl.Execute(buf, root)
	if err != nil {
		return "", err
	}
	return RemoveEmptyRow(buf.String()), nil
}

// 必须返回1-2个参数，其他一个必须是error并且放在最后一个返回值，至于sql返回有没有都行
// 默认只返回3种类型 slice，单个结构体，int float64 string
func (s *sqlInvoke) invokeSelect(local *context.LocalStack, args []reflect.Value, sqlEle *MapperElementXml) []reflect.Value {
	var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

	errorFlag := GetErrorHandleFlag(local) //0 panic 1 return
	con := GetDbConnection(local)
	if con == nil {
		var errortip string = "上下文中找不到数据库链接"
		if errorFlag == 0 {
			panic(errortip)
		} else {
			var defaultError *DaoException = &DaoException{exception.FrameException{Code: 505, Message: errortip}}
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(defaultError)}
			} else {
				return []reflect.Value{reflect.ValueOf(defaultError)}
			}
		}
	}

	sql, err1 := s.getSqlFromTpl(local, args, sqlEle)
	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	sqlParam, newsql, err2 := s.getArgumentsFromSql(local, args, sql)
	if err2 != nil {
		if errorFlag == 0 {
			panic(err2)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err2)}
			} else {
				return []reflect.Value{reflect.ValueOf(err2)}
			}
		}
	}
	if newsql != "" {
		sql = newsql
	}

	if s.Logger != nil && s.Logger.IsDebugEnable() {
		s.Logger.Debug(local, "--SQL %s: %s", sqlEle.Id, sql)
		s.Logger.Debug(local, "PARAM %s: %s", sqlEle.Id, GetSqlParamterStri(sqlParam))
	}

	stmt, err := con.Con.PrepareContext(con.Ctx, sql)
	if err != nil {
		if errorFlag == 0 {
			panic(err)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err)}
			} else {
				return []reflect.Value{reflect.ValueOf(err)}
			}
		}
	}
	defer stmt.Close()

	var queryResult *reflect.Value
	var queryError error
	switch s.returnSqlType.Kind() {
	case reflect.Slice:
		queryResult, queryError = s.selectList(local, sqlEle, stmt, sqlParam, errorFlag)
	case reflect.Ptr:
		queryResult, queryError = s.selectRow(local, sqlEle, stmt, sqlParam, errorFlag)
	case reflect.Int, reflect.Int64, reflect.Float64:
		queryResult, queryError = s.selectRow(local, sqlEle, stmt, sqlParam, errorFlag)
	default:
		fmt.Println("2")
	}

	if queryError != nil {
		if errorFlag == 0 {
			panic(queryError)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(queryError)}
			} else {
				return []reflect.Value{reflect.ValueOf(queryError)}
			}
		}
	}

	if s.returnSqlType != nil {
		return []reflect.Value{*queryResult, nilError}
	} else {
		return []reflect.Value{nilError}
	}

}

func (s *sqlInvoke) invokeUpdate(local *context.LocalStack, args []reflect.Value, sqlEle *MapperElementXml) []reflect.Value {
	var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

	errorFlag := GetErrorHandleFlag(local) //0 panic 1 return
	con := GetDbConnection(local)
	if con == nil {
		var errortip string = "上下文中找不到数据库链接"
		if errorFlag == 0 {
			panic(errortip)
		} else {
			var defaultError *DaoException = &DaoException{exception.FrameException{Code: 505, Message: errortip}}
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(defaultError)}
			} else {
				return []reflect.Value{reflect.ValueOf(defaultError)}
			}
		}
	}

	sql, err1 := s.getSqlFromTpl(local, args, sqlEle)
	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	sqlParam, newsql, err2 := s.getArgumentsFromSql(local, args, sql)
	if err2 != nil {
		if errorFlag == 0 {
			panic(err2)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err2)}
			} else {
				return []reflect.Value{reflect.ValueOf(err2)}
			}
		}
	}
	if newsql != "" {
		sql = newsql
	}
	//fmt.Printf("Sql[%s]: %s \n", sqlEle.Id, sql)
	//fmt.Printf("Paramters[%s]: %s \n", sqlEle.Id, GetSqlParamterStri(sqlParam))

	if s.Logger != nil && s.Logger.IsDebugEnable() {
		s.Logger.Debug(local, "--SQL %s: %s", sqlEle.Id, sql)
		s.Logger.Debug(local, "PARAM %s: %s", sqlEle.Id, GetSqlParamterStri(sqlParam))
	}

	stmt, err := con.Con.PrepareContext(con.Ctx, sql)
	if err != nil {
		if errorFlag == 0 {
			panic(err)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err)}
			} else {
				return []reflect.Value{reflect.ValueOf(err)}
			}
		}
	}
	defer stmt.Close()

	sqlResult, err1 := stmt.Exec(sqlParam...)

	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	if s.returnSqlType != nil {
		r1, _ := sqlResult.RowsAffected()

		if s.Logger != nil && s.Logger.IsDebugEnable() {
			s.Logger.Debug(local, "COUNT %s: %d", sqlEle.Id, int(r1))
		}

		return []reflect.Value{reflect.ValueOf(int(r1)), nilError}
	} else {
		return []reflect.Value{nilError}
	}
}

func (s *sqlInvoke) invokeDelete(local *context.LocalStack, args []reflect.Value, sqlEle *MapperElementXml) []reflect.Value {
	var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

	errorFlag := GetErrorHandleFlag(local) //0 panic 1 return
	con := GetDbConnection(local)
	if con == nil {
		var errortip string = "上下文中找不到数据库链接"
		if errorFlag == 0 {
			panic(errortip)
		} else {
			var defaultError *DaoException = &DaoException{exception.FrameException{Code: 505, Message: errortip}}
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(defaultError)}
			} else {
				return []reflect.Value{reflect.ValueOf(defaultError)}
			}
		}
	}

	sql, err1 := s.getSqlFromTpl(local, args, sqlEle)
	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	sqlParam, newsql, err2 := s.getArgumentsFromSql(local, args, sql)
	if err2 != nil {
		if errorFlag == 0 {
			panic(err2)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err2)}
			} else {
				return []reflect.Value{reflect.ValueOf(err2)}
			}
		}
	}
	if newsql != "" {
		sql = newsql
	}
	//fmt.Printf("Sql[%s]: %s \n", sqlEle.Id, sql)
	//fmt.Printf("Paramters[%s]: %s \n", sqlEle.Id, GetSqlParamterStri(sqlParam))
	if s.Logger != nil && s.Logger.IsDebugEnable() {
		s.Logger.Debug(local, "--SQL %s: %s", sqlEle.Id, sql)
		s.Logger.Debug(local, "PARAM %s: %s", sqlEle.Id, GetSqlParamterStri(sqlParam))
	}
	stmt, err := con.Con.PrepareContext(con.Ctx, sql)
	if err != nil {
		if errorFlag == 0 {
			panic(err)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err)}
			} else {
				return []reflect.Value{reflect.ValueOf(err)}
			}
		}
	}
	defer stmt.Close()

	sqlResult, err1 := stmt.Exec(sqlParam...)

	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	if s.returnSqlType != nil {
		r1, _ := sqlResult.RowsAffected()

		if s.Logger != nil && s.Logger.IsDebugEnable() {
			s.Logger.Debug(local, "COUNT %s: %d", sqlEle.Id, int(r1))
		}

		return []reflect.Value{reflect.ValueOf(int(r1)), nilError}
	} else {
		return []reflect.Value{nilError}
	}
}

func (s *sqlInvoke) invokeInsert(local *context.LocalStack, args []reflect.Value, sqlEle *MapperElementXml) []reflect.Value {
	var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

	errorFlag := GetErrorHandleFlag(local) //0 panic 1 return
	con := GetDbConnection(local)
	if con == nil {
		var errortip string = "上下文中找不到数据库链接"
		if errorFlag == 0 {
			panic(errortip)
		} else {
			var defaultError *DaoException = &DaoException{exception.FrameException{Code: 505, Message: errortip}}
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(defaultError)}
			} else {
				return []reflect.Value{reflect.ValueOf(defaultError)}
			}
		}
	}

	sql1, err1 := s.getSqlFromTpl(local, args, sqlEle)
	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	sqlParam, newsql, err2 := s.getArgumentsFromSql(local, args, sql1)
	if err2 != nil {
		if errorFlag == 0 {
			panic(err2)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err2)}
			} else {
				return []reflect.Value{reflect.ValueOf(err2)}
			}
		}
	}
	if newsql != "" {
		sql1 = newsql
	}
	//fmt.Printf("Sql[%s]: %s \n", sqlEle.Id, sql1)
	//fmt.Printf("Paramters[%s]: %s \n", sqlEle.Id, GetSqlParamterStri(sqlParam))

	if s.Logger != nil && s.Logger.IsDebugEnable() {
		s.Logger.Debug(local, "--SQL %s: %s", sqlEle.Id, sql1)
		s.Logger.Debug(local, "PARAM %s: %s", sqlEle.Id, GetSqlParamterStri(sqlParam))
	}

	stmt, err := con.Con.PrepareContext(con.Ctx, sql1)
	if err != nil {
		if errorFlag == 0 {
			panic(err)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err)}
			} else {
				return []reflect.Value{reflect.ValueOf(err)}
			}
		}
	}
	defer stmt.Close()

	sqlResult, err1 := stmt.Exec(sqlParam...)

	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			if s.returnSqlType != nil {
				return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err1)}
			} else {
				return []reflect.Value{reflect.ValueOf(err1)}
			}
		}
	}

	if s.baseMethod && s.tableDef != nil && s.tableDef.IdColumn != nil &&
		strings.ToUpper(s.tableDef.GenerationType) == "IDENTITY" {
		// 这种情况 args[1] 就是entity
		lid1, err3 := sqlResult.LastInsertId()

		if err3 != nil {
			if errorFlag == 0 {
				panic(err3)
			} else {
				if s.returnSqlType != nil {
					return []reflect.Value{*s.defaultReturnValue, reflect.ValueOf(err3)}
				} else {
					return []reflect.Value{reflect.ValueOf(err3)}
				}
			}
		}

		var llid1 interface{}
		switch s.tableDef.IdColumn.Field.Type.Kind() {
		case reflect.Int:
			llid1 = &sql.NullInt32{Int32: int32(lid1), Valid: true}
		case reflect.Int64:
			llid1 = &sql.NullInt64{Int64: lid1, Valid: true}
		}
		enptr := args[1].Elem().Elem()
		SetEntityFieldValue(&enptr, s.tableDef.IdColumn.Field, llid1)
	}

	if s.returnSqlType != nil {
		r1, _ := sqlResult.RowsAffected()
		if s.Logger != nil && s.Logger.IsDebugEnable() {
			s.Logger.Debug(local, "COUNT %s: %d", sqlEle.Id, int(r1))
		}
		return []reflect.Value{reflect.ValueOf(int(r1)), nilError}
	} else {
		return []reflect.Value{nilError}
	}
}

// getArgumentsFromSql 根据sql获取使用的变量值
//
// Paramter:
//
// - 使用变量值
//
// - 替换之后的sql
func (s *sqlInvoke) getArgumentsFromSql(local *context.LocalStack, args []reflect.Value, sql string) ([]interface{}, string, error) {
	// 去除局部变量参数
	if len(args) <= 1 {
		return nil, "", nil
	}

	// 只有一个参数 结构体 基础类型 string
	var params []string
	if s.providerConfig != nil && s.providerConfig.Param != "" {
		params = strings.Split(s.providerConfig.Param, ",")
	}
	var root interface{}
	if len(args) == 2 {
		if len(params) >= 2 && params[1] != "_" && params[1] != "" {
			root1 := make(map[string]interface{})
			root1[params[1]] = args[1].Interface()
			root = root1
		} else {
			root = args[1].Interface()
		}
	} else {
		root1 := make(map[string]interface{})
		for i := 1; i < len(params); i++ {
			if params[i] != "_" && params[i] != "" {
				root1[params[i]] = args[i].Interface()
			}
		}
		root = root1
	}

	variables, nsql := parseAndGetSqlVariables(sql)
	if len(variables) == 0 {
		return nil, "", nil
	}
	var result []interface{}
	for _, v := range variables {
		m := core.GetVariableValue(root, v, 1)
		result = append(result, m)
	}
	return result, nsql, nil
}

func (s *sqlInvoke) selectList(local *context.LocalStack, sqlEle *MapperElementXml, stmt *sql.Stmt, param []interface{}, errorFlag int) (*reflect.Value, error) {
	result, err1 := stmt.Query(param...)
	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			return nil, err1
		}
	}
	defer func() {
		if result != nil {
			result.Close() //可以关闭掉未scan连接一直占用
		}
	}()

	pageSize := 50
	queryCount := 0
	total := reflect.MakeSlice(s.returnSqlType, 0, 0)
	current := reflect.MakeSlice(s.returnSqlType, 0, pageSize)
	currentCount := 0

	if s.sqlFieldMap == nil {
		r1, e1 := result.ColumnTypes()
		if e1 == nil {
			sts := make([]*sqlColumnType, len(r1), len(r1))
			for k, ct := range r1 {
				var m1 *sqlColumnType = s.coverToGoType(ct)
				sts[k] = m1
			}
			s.sqlFieldMap = sts
		}
	}

	if s.Logger != nil && s.Logger.IsTraceEnable() {
		s.Logger.Trace(local, "column列类型 %s", GetSqlColumnType(s.sqlFieldMap))
	}

	for result.Next() {
		if queryCount != 0 && currentCount >= pageSize {
			total = reflect.AppendSlice(total, current)
			current = reflect.MakeSlice(s.returnSqlType, 0, pageSize)
			currentCount = 0
		}

		r1, err2 := s.scanRow(result)
		if err2 != nil {
			if errorFlag == 0 {
				panic(err2)
			} else {
				return nil, err2
			}
		}
		//fmt.Println(r1.Interface())
		current = reflect.Append(current, *r1)

		queryCount++
		currentCount++
	}

	if s.Logger != nil && s.Logger.IsDebugEnable() {
		s.Logger.Debug(local, "COUNT %s: %d", sqlEle.Id, queryCount)
	}

	if queryCount > 0 {
		total = reflect.AppendSlice(total, current)
		total = total.Slice(0, queryCount)
	}
	return &total, nil
}

func (s *sqlInvoke) selectRow(local *context.LocalStack, sqlEle *MapperElementXml, stmt *sql.Stmt, param []interface{}, errorFlag int) (*reflect.Value, error) {
	result, err1 := stmt.Query(param...)
	if err1 != nil {
		if errorFlag == 0 {
			panic(err1)
		} else {
			return nil, err1
		}
	}
	defer func() {
		if result != nil {
			result.Close() //可以关闭掉未scan连接一直占用
		}
	}()

	if s.sqlFieldMap == nil {
		r1, e1 := result.ColumnTypes()
		if e1 == nil {
			sts := make([]*sqlColumnType, len(r1), len(r1))
			for k, ct := range r1 {
				var m1 *sqlColumnType = s.coverToGoType(ct)
				sts[k] = m1
			}
			s.sqlFieldMap = sts
		}
	}

	if s.Logger != nil && s.Logger.IsTraceEnable() {
		s.Logger.Trace(local, "column列类型 %s", GetSqlColumnType(s.sqlFieldMap))
	}

	if result.Next() {
		r1, err2 := s.scanRow(result)
		if err2 != nil {
			if errorFlag == 0 {
				panic(err2)
			} else {
				return nil, err2
			}
		}
		if s.Logger != nil && s.Logger.IsDebugEnable() {
			s.Logger.Debug(local, "COUNT %s: %d", sqlEle.Id, 1)
		}
		return r1, nil
	} else {
		if s.Logger != nil && s.Logger.IsDebugEnable() {
			s.Logger.Debug(local, "COUNT %s: %d", sqlEle.Id, 0)
		}
		return s.defaultReturnValue, nil
	}
}

func (s *sqlInvoke) scanRow(result *sql.Rows) (*reflect.Value, error) {
	valueptr := make([]interface{}, len(s.sqlFieldMap), len(s.sqlFieldMap))
	for k, v := range s.sqlFieldMap {
		d1 := GetSqlFieldReturnDefaultValue(v.defaultType)
		//fmt.Println(v.field.Name, v.defaultType.Kind(), reflect.ValueOf(d1).Elem().Kind())
		valueptr[k] = d1
	}
	e1 := result.Scan(valueptr...) //不scan会导致连接不释放
	if e1 != nil {
		return nil, e1
	}
	//fmt.Println(valueptr[0])
	var result1 *reflect.Value

	//fmt.Println(s.returnSqlElementType.Kind(),s.returnSqlElementType.String())
	switch s.returnSqlElementType.Kind() {
	case reflect.Struct:
		hp := reflect.New(s.returnSqlElementType)
		hv := hp.Elem()
		for k, v := range s.sqlFieldMap {
			if v.field != nil {
				SetEntityFieldValue(&hv, v.field, valueptr[k])
			}
		}
		result1 = &hp
	case reflect.String, reflect.Int, reflect.Int64, reflect.Float64:
		hp := GetRowColumnValue(s.sqlFieldMap[0].defaultType, valueptr[0])
		if hp == nil {
			zv := reflect.Zero(s.returnSqlElementType)
			result1 = &zv
		} else {
			result1 = hp
		}
	}
	return result1, nil
}

func (s *sqlInvoke) coverToGoType(ct *sql.ColumnType) *sqlColumnType {
	result := sqlColumnType{column: ct}
	addDefaultType := true

	if len(s.structFieldMap) > 0 {
		columnName := ct.Name()
		// AddId-->Addid有问题
		fieldName := core.GetCamelCaseName(columnName)
		if field, ok := s.structFieldMap[fieldName]; ok {
			addDefaultType = false
			result.field = &field
			result.defaultType = field.Type
		}
	}

	// 单列返回
	if addDefaultType {
		switch s.returnSqlElementType.Kind() {
		case reflect.Int64:
			result.defaultType = reflect.TypeOf(int64(1))
			addDefaultType = false
		case reflect.Int:
			result.defaultType = reflect.TypeOf(int(1))
			addDefaultType = false
		case reflect.Float64:
			result.defaultType = reflect.TypeOf(float64(1))
			addDefaultType = false
		}
	}

	if addDefaultType {
		databasetype := strings.ToLower(ct.DatabaseTypeName())
		if strings.Index(databasetype, "int") >= 0 {
			result.defaultType = reflect.TypeOf(int(1))
		} else if strings.Index(databasetype, "decimal") >= 0 {
			result.defaultType = reflect.TypeOf(float64(1.0))
		} else if strings.Index(databasetype, "char") >= 0 {
			result.defaultType = reflect.TypeOf(string(""))
		} else if strings.Index(databasetype, "date") >= 0 {
			n := time.Now()
			result.defaultType = reflect.TypeOf(&n)
		} else {
			result.defaultType = reflect.TypeOf(string(""))
		}
	}
	return &result
}

func newSqlInvoke(
	target interface{}, //对象
	clazz *proxyclass.ProxyClass, //代理信息
	method *proxyclass.ProxyMethod, //当前方法代理信息
	mapper map[string]*MapperElementXml, //对应sql节点
	returnSqlType reflect.Type, //返回的类型 不是error 如果没有就nil
	providerConfig *SqlProviderConfig,
	defaultReturnValue *reflect.Value, //默认返回类型值
	structFieldMap map[string]reflect.StructField,
	entityptr interface{},
	baseMethod bool,
	//base xml 节点
	baseXmlEle map[string]*MapperElementXml,
	//entity 解析结构
	tableDef *TableDef, local *context.LocalStack, applicationContext *application.FrameApplicationContext,
) *sqlInvoke {

	var returnSqlElementType reflect.Type = nil
	if returnSqlType != nil {
		switch returnSqlType.Kind() {
		case reflect.Slice:
			if baseMethod {
				// base find
				entityType := reflect.TypeOf(entityptr)

				//returnSqlType = reflect.SliceOf(entityType)

				returnSqlElementType = entityType.Elem()

				defaultReturnValue = core.GetMethodReturnDefaultValue(returnSqlType)
				structFieldMap = core.GetStructField(entityType)
			} else {
				if returnSqlType.Elem().Kind() == reflect.Ptr {
					returnSqlElementType = returnSqlType.Elem().Elem()
				} else {
					returnSqlElementType = returnSqlType.Elem()
				}
			}
		case reflect.Ptr:
			if returnSqlType.Elem().Kind() == reflect.Struct {
				returnSqlElementType = returnSqlType.Elem()
			} else {
				returnSqlElementType = returnSqlType
			}
		case reflect.Interface:
			//base Get
			returnSqlType = reflect.TypeOf(entityptr)
			returnSqlElementType = returnSqlType.Elem()

			defaultReturnValue = core.GetMethodReturnDefaultValue(returnSqlType)
			structFieldMap = core.GetStructField(returnSqlType)
		default:
			returnSqlElementType = returnSqlType
		}
	}
	var logger logclass.AppLoger
	if applicationContext != nil && applicationContext.LogFactory != nil {
		logger = applicationContext.LogFactory.GetLoggerType(reflect.TypeOf(target).Elem())
	}
	return &sqlInvoke{
		Logger:               logger,
		target:               target,
		clazz:                clazz,
		method:               method,
		mapper:               mapper,
		providerConfig:       providerConfig,
		returnSqlType:        returnSqlType,
		defaultReturnValue:   defaultReturnValue,
		structFieldMap:       structFieldMap,
		returnSqlElementType: returnSqlElementType,
		baseXmlEle:           baseXmlEle,
		tableDef:             tableDef,
		baseMethod:           baseMethod,
	}
}

func GetDaoConfig(target1 proxyclass.ProxyTarger) *DaoConfig {
	if target1 == nil || target1.ProxyTarget() == nil || len(target1.ProxyTarget().Annotations) == 0 {
		return nil
	}
	for _, anno := range target1.ProxyTarget().Annotations {
		if anno.Name == proxyclass.AnnotationDao {
			if v, f := anno.Value[AnnotationDaoConfigValueKey]; f {
				configDao := v.(*DaoConfig)
				return configDao
			}
		}
	}
	return nil
}

type FrameOrmFactory struct {
	logger logclass.AppLoger
	lock   sync.Mutex
}

func (f *FrameOrmFactory) intiFrameOrmFactory(local *context.LocalStack, applicationContext *application.FrameApplicationContext) {
	if f.logger != nil {
		return
	}
	f.lock.Lock()
	defer func() {
		f.lock.Unlock()
	}()
	if f.logger != nil {
		return
	}
	f.logger = applicationContext.LogFactory.GetLoggerType(reflect.TypeOf(f).Elem())
	f.logger.Debug(local, "[初始化] FrameOrmFactory")
}

func (f *FrameOrmFactory) AddMapperProxyTarget(local *context.LocalStack, target1 proxyclass.ProxyTarger,
	applicationContext *application.FrameApplicationContext) {
	f.intiFrameOrmFactory(local, applicationContext)
	//解析字段方法 包裹一层
	rv := reflect.ValueOf(target1)
	rt := rv.Elem().Type()

	var tableDef *TableDef
	daoConfig := GetDaoConfig(target1)

	daoEntity := daoConfig.Entity
	if daoEntity != nil {
		tableDef = parseEntityType(daoEntity)
	}

	xmlele := mapperFactory.ParseXml(target1, daoConfig.Xml, tableDef)
	var baseXmlEle map[string]*MapperElementXml

	if BaseXml != "" {
		buf := &bytes.Buffer{}
		err1 := BaseXmlTpl.Execute(buf, tableDef)
		if err1 != nil {
			panic(err1)
		}
		baseXmlEle = mapperFactory.ParseXml(target1, buf.String(), tableDef)
	}

	methodRef := make(map[string]*proxyclass.ProxyMethod)
	if len(target1.ProxyTarget().Methods) != 0 {
		for _, md := range target1.ProxyTarget().Methods {
			methodRef[md.Name] = md
		}
	}

	if m1 := rt.NumField(); m1 > 0 {
		for i := 0; i < m1; i++ {
			field := rt.Field(i)
			if field.Type == BaseDaoType {
				if m2 := field.Type.NumField(); m2 > 0 {
					for j := 0; j < m2; j++ {
						basefield := field.Type.Field(j)
						target := rv.Elem().FieldByName(field.Name)
						f.logger.Debug(local, "[初始化] dao的方法初始化 实例 %s 方法 %s",
							util.ClassUtil.GetJavaClassNameByType(rt), basefield.Name)
						f.addCallerToField(target1, &target, &basefield, methodRef,
							baseXmlEle, daoConfig, true,
							baseXmlEle, tableDef, local, applicationContext)
					}
				}
			} else if field.Type.Kind() == reflect.Func && rv.Elem().FieldByName(field.Name).IsNil() {
				target := rv.Elem()
				f.logger.Debug(local, "[初始化] dao的方法初始化 实例 %s 方法 %s",
					util.ClassUtil.GetJavaClassNameByType(rt), field.Name)
				f.addCallerToField(target1, &target, &field, methodRef,
					xmlele, daoConfig, false,
					baseXmlEle, tableDef, local, applicationContext)
			}
		}
	}
	core.AddClassProxy(target1)
}

func (f *FrameOrmFactory) addCallerToField(target1 proxyclass.ProxyTarger,
	target *reflect.Value,
	field *reflect.StructField,
	methodRef map[string]*proxyclass.ProxyMethod,
	xmlele map[string]*MapperElementXml,
	entityptr interface{},
	baseMethod bool,
	baseXmlEle map[string]*MapperElementXml,
	tableDef *TableDef, local *context.LocalStack, applicationContext *application.FrameApplicationContext,
) {
	call := target.FieldByName(field.Name)

	methodName := strings.ReplaceAll(field.Name, "_", "")
	methodSetting, ok := methodRef[methodName]
	if !ok {
		methodSetting = &proxyclass.ProxyMethod{Name: methodName}
	}

	var providerConfig *SqlProviderConfig = nil
	if methodSetting != nil && len(methodSetting.Annotations) > 0 {
		for _, anno := range methodSetting.Annotations {
			if anno.Name == AnnotationSqlProviderConfig {
				if provider, f := anno.Value[AnnotationSqlProviderConfigValueKey]; f {
					providerConfig = provider.(*SqlProviderConfig)
				}
			}
		}
	}

	var invoker *sqlInvoke
	fo := field.Type.NumOut()
	if fo >= 2 {
		defaultReturnValue := core.GetMethodReturnDefaultValue(field.Type.Out(0))
		structFields := core.GetStructField(field.Type.Out(0))
		invoker = newSqlInvoke(target1, target1.ProxyTarget(),
			methodSetting,
			xmlele, field.Type.Out(0),
			providerConfig, defaultReturnValue, structFields, entityptr, baseMethod,
			baseXmlEle,
			tableDef, local, applicationContext)
	} else {
		invoker = newSqlInvoke(target1, target1.ProxyTarget(), methodSetting,
			xmlele, nil,
			providerConfig, nil, nil, entityptr, baseMethod,
			baseXmlEle,
			tableDef, local, applicationContext)
	}

	proxyCall := func(command *sqlInvoke) reflect.Value {
		newCall := reflect.MakeFunc(field.Type, func(in []reflect.Value) []reflect.Value {
			return command.invoke(in[0].Interface().(*context.LocalStack), in)
		})
		return newCall
	}(invoker)
	call.Set(proxyCall)
}

var frameOrmFactory FrameOrmFactory = FrameOrmFactory{}

func GetFrameOrmFactory() *FrameOrmFactory {
	return &frameOrmFactory
}

func NewSqlProvierConfigAnnotation(param string) *proxyclass.AnnotationClass {
	return &proxyclass.AnnotationClass{
		Name: AnnotationSqlProviderConfig,
		Value: map[string]interface{}{
			AnnotationSqlProviderConfigValueKey: &SqlProviderConfig{
				Param: param,
			},
		},
	}
}

// parseAndGetSqlVariables #{ada} 获取ada
func parseAndGetSqlVariables(sql string) ([]string, string) {
	reg := regexp.MustCompile(`(?m)#\{(\S+?)\}`)
	result := reg.FindAllStringSubmatch(sql, -1)
	if len(result) > 0 {
		var r1 []string
		for _, k := range result {
			r1 = append(r1, k[1])
		}
		return r1, reg.ReplaceAllString(sql, "?")
	} else {
		return nil, ""
	}
}

// GetSqlFieldReturnDefaultValue  用来接受sql column 返回的值 类型 int int64 float64 string *Time
// 当有错误的时候 返回这个默认结果的指针 和 错误
func GetSqlFieldReturnDefaultValue(rtType reflect.Type) interface{} {
	switch rtType.Kind() {
	case reflect.String:
		v := sql.NullString{}
		return &v
	case reflect.Int64:
		v := sql.NullInt64{}
		return &v
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		v := sql.NullInt32{}
		return &v
	case reflect.Float32, reflect.Float64:
		v := sql.NullFloat64{}
		return &v
	case reflect.Ptr:
		eleType := rtType.Elem()
		if eleType == GoTimeType {
			v := sql.NullTime{}
			return &v
		} else if eleType == SqlNullStringType {
			v := sql.NullString{}
			return &v
		} else if eleType == SqlNullTimeType {
			v := sql.NullTime{}
			return &v
		} else if eleType == SqlNullInt64Type {
			v := sql.NullInt64{}
			return &v
		} else if eleType == SqlNullInt32Type {
			v := sql.NullInt32{}
			return &v
		} else if eleType == SqlNullFloat64Type {
			v := sql.NullFloat64{}
			return &v
		} else if eleType == SqlNullBoolType {
			v := sql.NullBool{}
			return &v
		}
		return nil
	default:
		panic(fmt.Sprintf("%s找不到对应默认值", rtType.String()))
	}
}

// GetRowColumnValue 将查询出来的column值转换成golang类型 value sql.null* 的指针
// 如果go类型是基础类型或者string 如果没有值就返回nil， 如果返回都是指针 那么只能用于struct里面的field
func GetRowColumnValue(ty reflect.Type, value interface{}) *reflect.Value {
	switch ty.Kind() {
	case reflect.String:
		s := value.(*sql.NullString)
		if s.Valid {
			n := reflect.ValueOf(s.String)
			return &n
		} else {
			return nil
		}
	case reflect.Int64:
		s := value.(*sql.NullInt64)
		if s.Valid {
			n := reflect.ValueOf(s.Int64)
			return &n
		} else {
			return nil
		}
	case reflect.Int:
		s := value.(*sql.NullInt32)
		if s.Valid {
			n := reflect.ValueOf(int(s.Int32))
			return &n
		} else {
			return nil
		}
	case reflect.Float64:
		s := value.(*sql.NullFloat64)
		if s.Valid {
			n := reflect.ValueOf(s.Float64)
			return &n
		} else {
			return nil
		}
	case reflect.Ptr:
		//fix field is *Time ,same as ptr
		eleType := ty.Elem()
		if eleType == GoTimeType {
			s := value.(*sql.NullTime)
			if s.Valid {
				n := reflect.ValueOf(&(s.Time))
				return &n
			}
		} else {
			switch p1 := value.(type) {
			case *sql.NullString:
				if p1.Valid {
					n := reflect.ValueOf(p1)
					return &n
				}
			case *sql.NullTime:
				if p1.Valid {
					n := reflect.ValueOf(p1)
					return &n
				}
			case *sql.NullInt64:
				if p1.Valid {
					n := reflect.ValueOf(p1)
					return &n
				}
			case *sql.NullInt32:
				if p1.Valid {
					n := reflect.ValueOf(p1)
					return &n
				}
			case *sql.NullFloat64:
				if p1.Valid {
					n := reflect.ValueOf(p1)
					return &n
				}
			}
		}
		return nil
	}
	panic(fmt.Errorf("%s 找不到对应处理类型", ty.String()))
}

// SetEntityFieldValue target 目标对象  name fieldname value sql.null* 的指针
// 如果value指针对应的是struct 根据ptr来判断是否是指针类型
func SetEntityFieldValue(target *reflect.Value, field *reflect.StructField, value interface{}) {
	rv := GetRowColumnValue(field.Type, value)
	if rv != nil {
		(*target).FieldByName(field.Name).Set(*rv)
	}
	//switch field.Type.Kind() {
	//case reflect.String:
	//	s := value.(*sql.NullString)
	//	if s.Valid {
	//		(*target).FieldByName(field.Name).Set(reflect.ValueOf(s.String))
	//	}
	//case reflect.Int64:
	//	s := value.(*sql.NullInt64)
	//	if s.Valid {
	//		(*target).FieldByName(field.Name).Set(reflect.ValueOf(s.Int64))
	//	}
	//case reflect.Int:
	//	s := value.(*sql.NullInt32)
	//	if s.Valid {
	//		(*target).FieldByName(field.Name).Set(reflect.ValueOf(int(s.Int32)))
	//	}
	//case reflect.Float64:
	//	s := value.(*sql.NullFloat64)
	//	if s.Valid {
	//		(*target).FieldByName(field.Name).Set(reflect.ValueOf(s.Float64))
	//	}
	//case reflect.Ptr:
	//	//fix field is *Time ,same as ptr
	//	eleType := field.Type.Elem()
	//	if eleType == GoTimeType {
	//		s := value.(*sql.NullTime)
	//		if s.Valid {
	//			(*target).FieldByName(field.Name).Set(reflect.ValueOf(&(s.Time)))
	//		}
	//	} else if eleType == SqlNullStringType {
	//		s := value.(*sql.NullString)
	//		if s.Valid {
	//			(*target).FieldByName(field.Name).Set(reflect.ValueOf(s))
	//		}
	//	} else if eleType == SqlNullTimeType {
	//		s := value.(*sql.NullTime)
	//		if s.Valid {
	//			(*target).FieldByName(field.Name).Set(reflect.ValueOf(s))
	//		}
	//	} else if eleType == SqlNullInt64Type {
	//		s := value.(*sql.NullInt64)
	//		if s.Valid {
	//			(*target).FieldByName(field.Name).Set(reflect.ValueOf(s))
	//		}
	//	} else if eleType == SqlNullInt32Type {
	//		s := value.(*sql.NullInt32)
	//		if s.Valid {
	//			(*target).FieldByName(field.Name).Set(reflect.ValueOf(s))
	//		}
	//	} else if eleType == SqlNullFloat64Type {
	//		s := value.(*sql.NullFloat64)
	//		if s.Valid {
	//			(*target).FieldByName(field.Name).Set(reflect.ValueOf(s))
	//		}
	//	}
	//}
}

func GetSqlNullTypeValue(p interface{}) interface{} {
	if p == nil {
		return nil
	}
	if reflect.ValueOf(p).IsZero() {
		return nil
	}
	switch p1 := p.(type) {
	case *sql.NullString:
		return p1.String
	case *sql.NullInt64:
		return p1.Int64
	case *sql.NullInt32:
		return int(p1.Int32)
	case *sql.NullFloat64:
		return p1.Float64
	case *sql.NullTime:
		return p1.Time
	case *sql.NullBool:
		return p1.Bool
	}
	return p
}

func GetSqlParamterStri(values []interface{}) string {
	if len(values) == 0 {
		return ""
	}
	var result strings.Builder
	for _, v := range values {
		result.WriteString(fmt.Sprint(v))
		result.WriteString(" ")
	}
	return result.String()
}
func GetSqlColumnType(values []*sqlColumnType) string {
	if len(values) == 0 {
		return ""
	}
	var result strings.Builder
	for _, v := range values {
		result.WriteString(fmt.Sprintf("column %s %s fieldType %s", v.column.Name(), v.column.DatabaseTypeName(), v.defaultType.String()))
		result.WriteString(" ")
	}
	return result.String()
}

//reg := regexp.MustCompile(`(?m)(^\s+|\s+$)`)
var removeEmptyRowReg *regexp.Regexp = regexp.MustCompile(`(?m)^\s*$\n`)

func RemoveEmptyRow(content string) string {
	return removeEmptyRowReg.ReplaceAllString(content, "")
}

func GetDaoAnnotation(annotations []*proxyclass.AnnotationClass) *proxyclass.AnnotationClass {
	for _, annotation := range annotations {
		if annotation.Name == proxyclass.AnnotationDao {
			return annotation
		}
	}
	return nil
}

func NewDaoAnnotation(xml string, entity interface{}) *proxyclass.AnnotationClass {
	return &proxyclass.AnnotationClass{
		Name: proxyclass.AnnotationDao,
		Value: map[string]interface{}{
			AnnotationDaoConfigValueKey: &DaoConfig{
				Xml:    xml,
				Entity: entity,
			},
		},
	}
}
