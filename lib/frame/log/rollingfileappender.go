package log

import (
	"bufio"
	"context"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"io"
	"os"
	"time"
)

type RollingFileAppenderImpl struct {
	Property   *AppenderProperty
	Target     io.Writer
	FilePath   string
	FileTarget *os.File
	FileBuffer *bufio.Writer

	TimePattern     string
	FileNamePattern string
	RollingRule     string
	BaseAppender
}

func (r *RollingFileAppenderImpl) AppendRow(local context.Context, level string,
	config *LoggerConfig, row string, err interface{}) {
	if r.BaseAppender.IsAppendRow(local, level, config, r.Property) {
		r.Property.Layout.DoLayout(local, level, config, row, err)
		r.FileBuffer.Flush()
	}
}

func (r *RollingFileAppenderImpl) NewAppender(ele *LogAppenderXmlEle) LogAppender {

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
	r.BaseAppender.SetAppender(ele, bufferWriter, result)
	return LogAppender(result)

}

func (r *RollingFileAppenderImpl) SetAppenderProperty(property *AppenderProperty) {
	r.Property = property
}

func (r *RollingFileAppenderImpl) AppenderKey() string {
	return RollingFileAppenderAdapterKey
}
