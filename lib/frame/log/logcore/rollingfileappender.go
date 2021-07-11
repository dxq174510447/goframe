package logcore

import (
	"bufio"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"io"
	"os"
	"time"
)

type RollingFileAppenderImpl struct {
	Property   *logclass.AppenderProperty
	Target     io.Writer
	FilePath   string
	FileTarget *os.File
	FileBuffer *bufio.Writer

	TimePattern     string
	FileNamePattern string
	RollingRule     string
}

func (r *RollingFileAppenderImpl) AppendRow(local *context.LocalStack, level string, config *logclass.LoggerConfig, row string, err interface{}) {
	if IsAppendRow(local, level, config, r.Property) {
		r.Property.Layout.DoLayout(local, level, config, row, err)
		r.FileBuffer.Flush()
	}
}

func (r *RollingFileAppenderImpl) NewAppender(ele *logclass.LogAppenderXmlEle) logclass.LogAppender {

	namePattern := ele.RollingPolicy.FileNamePattern
	timePattern := GetTimePatternFromFileNamePattern(namePattern)

	var rollingRule string

	now := time.Now()
	nowStr := util.DateUtil.FormatByType(&now, timePattern)
	for _, rule := range LogRollingType {
		end := GetRollingRule(rule, &now, 1)
		endStr := util.DateUtil.FormatByType(end, timePattern)
		if nowStr != "" && endStr != "" && nowStr != endStr {
			rollingRule = rule
			break
		}
	}

	fileTarget, err := os.OpenFile(ele.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	bufferWriter := bufio.NewWriter(fileTarget)

	//layout := NewLayout(ele.Encoder[0].Pattern, bufferWriter)
	result := &RollingFileAppenderImpl{
		Target:          bufferWriter,
		FileTarget:      fileTarget,
		FilePath:        ele.File,
		FileBuffer:      bufferWriter,
		TimePattern:     timePattern,
		FileNamePattern: namePattern,
		RollingRule:     rollingRule,
	}
	SetAppender(ele, bufferWriter, result)
	return logclass.LogAppender(result)

}

func (r *RollingFileAppenderImpl) SetAppenderProperty(property *logclass.AppenderProperty) {
	r.Property = property
}

func (r *RollingFileAppenderImpl) AppenderKey() string {
	return RollingFileAppenderAdapterKey
}
