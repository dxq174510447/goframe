package dbcore

import (
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type TableDef struct {
	Name           string
	DbName         string
	GenerationType string
	IdColumn       *TableColumnDef
	Columns        []*TableColumnDef
}

type TableColumnDef struct {
	FieldName  string
	Field      *reflect.StructField
	ColumnName string
	Transient  bool
	Updatable  bool
}

func fieldNameToColumanName(fieldName string) string {
	var fields []string

	reg := regexp.MustCompile(`[A-Z]`)
	result := reg.FindAllIndex([]byte(fieldName), -1)

	position := 0
	n := len(result)
	if n > 0 {
		for i := 0; i < n; i++ {
			if i == 0 && result[i][0] == 0 {
				continue
			}
			fields = append(fields, fieldName[position:result[i][0]])
			position = result[i][0]
		}
		if position <= (len(fieldName) - 1) {
			fields = append(fields, fieldName[position:])
			if position == (len(fieldName) - 1) {
				fields = append(fields, "")
			}
		}
		return strings.ToLower(strings.Join(fields, "_"))
	}
	return strings.ToLower(fieldName)
}

func parseEntityType(entity interface{}) *TableDef {
	ty := reflect.TypeOf(entity).Elem()
	n := ty.NumField()

	var td TableDef
	for i := 0; i < n; i++ {

		var tc *TableColumnDef = &TableColumnDef{}

		field := ty.Field(i)
		tc.FieldName = field.Name
		tc.Field = &field

		if column, ok := field.Tag.Lookup("column"); ok {
			if column != "" {
				tc.ColumnName = column
			}
		}

		if tc.ColumnName == "" {
			//fmt.Println(fieldNameToColumanName("a"))
			//fmt.Println(fieldNameToColumanName("abc"))
			//fmt.Println(fieldNameToColumanName("abcDe"))
			//fmt.Println(fieldNameToColumanName("abcDeFGh"))
			//fmt.Println(fieldNameToColumanName("abcDeFGhJ"))
			//
			//fmt.Println(fieldNameToColumanName("A"))
			//fmt.Println(fieldNameToColumanName("Abc"))
			//fmt.Println(fieldNameToColumanName("AbcDe"))
			//fmt.Println(fieldNameToColumanName("AbcDeFGh"))
			//fmt.Println(fieldNameToColumanName("AbcDeFGhJ"))
			tc.ColumnName = fieldNameToColumanName(tc.FieldName)
		}

		if id, ok := field.Tag.Lookup("id"); ok {
			if id != "" {
				td.IdColumn = tc
			}
		}

		tc.Transient = false
		if transient, ok := field.Tag.Lookup("transient"); ok {
			if transient == "true" {
				tc.Transient = true
			}
		}

		tc.Updatable = true
		if updatable, ok := field.Tag.Lookup("updatable"); ok {
			if updatable == "false" {
				tc.Updatable = false
			}
		}
		if generationType, ok := field.Tag.Lookup("GenerationType"); ok {
			if generationType != "" {
				td.GenerationType = generationType
			}
		}
		if dbname, ok := field.Tag.Lookup("dbname"); ok {
			if dbname != "" {
				td.DbName = dbname
			}
		}
		if table, ok := field.Tag.Lookup("table"); ok {
			if table != "" {
				td.Name = table
			}
		}

		if !tc.Transient {
			td.Columns = append(td.Columns, tc)
		}
	}

	if td.IdColumn != nil {
		sort.Slice(td.Columns, func(i, j int) bool {
			if td.Columns[i].FieldName == td.IdColumn.FieldName {
				return false
			}
			return false
		})
	}
	return &td
}
