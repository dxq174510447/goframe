package logcore

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"text/template"
)

type ConsoleAppenderImpl struct {
}

type LogFactory struct {
	//Appender map[string]LogAppender
}

func (l *LogFactory) Parse(content string, funcMap template.FuncMap) {

	var tpl *template.Template
	if funcMap == nil || len(funcMap) == 0 {
		tpl = template.Must(template.New(fmt.Sprintf("%s-logcore", util.DateUtil.FormatNowByType(util.DatePattern2))).Parse(content))
	} else {
		tpl = template.Must(template.New(fmt.Sprintf("%s-logcore", util.DateUtil.FormatNowByType(util.DatePattern2))).Funcs(funcMap).Parse(content))
	}

	buf := &bytes.Buffer{}
	param := make(map[string]interface{})
	err := tpl.Execute(buf, param)
	if err != nil {
		panic(err)
	}
	xml1 := util.StringUtil.RemoveEmptyRow(buf.String())

	config := &logclass.LogXmlEle{}

	err1 := xml.Unmarshal([]byte(xml1), config)

	if err1 != nil {
		panic(err1)
	}

	fmt.Println(util.JsonUtil.To2String(config))
	fmt.Println(config.Appender[0].Encoder[0].Pattern)
}
