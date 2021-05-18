package dbcore

import (
	"text/template"
)

const BaseXml = `

<?xml version="1.0" encoding="UTF-8" ?>
<mapper>
	<insert id="Save">
			insert into {{.Name}}(
			{{range $index, $ele := $.Columns}}{{if $index}},{{end}}{{$ele.ColumnName}}{{end}}
			) values (
			{{range $index, $ele := $.Columns}}{{if $index}},{{end}}#{{"{"}}{{$ele.FieldName}}{{"}"}}{{end}}
			)
	</insert>

	<update id="Update">
		update {{.Name}} 
		set
		{{range $index, $ele := $.Columns}}{{if $ele.Updatable}}{{if $index}},{{end}}{{$ele.ColumnName}} = #{{"{"}}{{$ele.FieldName}}{{"}"}}{{end}}{{end}}
		where {{.IdColumn.ColumnName}} = #{{"{"}}{{.IdColumn.FieldName}}{{"}"}}
	</update>

	<delete id="Delete">
		delete from {{.Name}} where {{.IdColumn.ColumnName}} = #{{"{"}}{{.IdColumn.FieldName}}{{"}"}}
	</delete>

	<select id="Get">
		select {{range $index, $ele := $.Columns}}{{if $index}},{{end}}{{$ele.ColumnName}}{{end}}
		from {{.Name}}
		where {{.IdColumn.ColumnName}} = #{{"{"}}{{.IdColumn.FieldName}}{{"}"}}
	</select>

	<select id="Find">
		select {{range $index, $ele := $.Columns}}{{if $index}},{{end}}{{$ele.ColumnName}}{{end}}
		from {{.Name}}
		where 1=1 
		{{range $index, $ele := $.Columns}}
		{{"{{if ."}}{{$ele.FieldName}}{{"}}"}}
		and {{$ele.ColumnName}} = #{{"{"}}{{$ele.FieldName}}{{"}"}}
		{{"{{end}}"}}
		{{end}}
	</select>

</mapper>

`

var BaseXmlTpl *template.Template = template.Must(template.New("baseSqlTpl").Parse(BaseXml))
