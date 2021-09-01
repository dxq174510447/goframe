package log

import (
	"context"
	"io"
	"os"
)

type ConsoleAppenderImpl struct {
	Property *AppenderProperty
	Target   io.Writer
	BaseAppender
}

func (c *ConsoleAppenderImpl) AppenderKey() string {
	return ConsoleAppenderAdapterKey
}

func (c *ConsoleAppenderImpl) NewAppender(ele *LogAppenderXmlEle) LogAppender {
	//layout := NewLayout(ele.Encoder[0].Pattern, os.Stdout)
	result := &ConsoleAppenderImpl{
		Target: os.Stdout,
	}
	c.BaseAppender.SetAppender(ele, os.Stdout, result)
	return LogAppender(result)
}

func (c *ConsoleAppenderImpl) SetAppenderProperty(property *AppenderProperty) {
	c.Property = property
}

func (c *ConsoleAppenderImpl) AppendRow(local context.Context, level string, config *LoggerConfig, row string, err interface{}) {
	if c.BaseAppender.IsAppendRow(local, level, config, c.Property) {
		c.Property.Layout.DoLayout(local, level, config, row, err)
		os.Stdout.Sync()
	}
}
