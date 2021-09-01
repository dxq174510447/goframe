package log

import (
	"bufio"
	"context"
	"io"
	"os"
)

type FileAppenderImpl struct {
	Property   *AppenderProperty
	Target     io.Writer
	FilePath   string
	FileTarget *os.File
	FileBuffer *bufio.Writer
	BaseAppender
}

func (f *FileAppenderImpl) AppendRow(local context.Context,
	level string,
	config *LoggerConfig, row string, err interface{}) {
	if f.BaseAppender.IsAppendRow(local, level, config, f.Property) {
		f.Property.Layout.DoLayout(local, level, config, row, err)
		f.FileBuffer.Flush()
	}
}

func (f *FileAppenderImpl) AppenderKey() string {
	return FileAppenderAdapterKey
}

func (f *FileAppenderImpl) SetAppenderProperty(property *AppenderProperty) {
	f.Property = property
}

func (f *FileAppenderImpl) NewAppender(ele *LogAppenderXmlEle) LogAppender {

	fileTarget, err := os.OpenFile(ele.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	bufferWriter := bufio.NewWriter(fileTarget)

	//layout := NewLayout(ele.Encoder[0].Pattern, bufferWriter)
	result := &FileAppenderImpl{
		Target:     bufferWriter,
		FileTarget: fileTarget,
		FilePath:   ele.File,
		FileBuffer: bufferWriter,
	}
	f.BaseAppender.SetAppender(ele, bufferWriter, result)
	return LogAppender(result)
}
