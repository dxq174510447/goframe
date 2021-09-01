package log

import (
	"context"
	"io"
	"strings"
	"sync"
)

type BaseAppender struct {
}

//IsAppendRow 是否过滤掉日志
func (b *BaseAppender) IsAppendRow(local context.Context, level string,
	config *LoggerConfig,
	property *AppenderProperty) bool {
	if len(property.Filter) == 0 {
		return true
	}
	for _, filter := range property.Filter {
		m := filter.LogDecide(local, level, config)
		m = strings.ToUpper(m)
		if m == DENYFilterReplay {
			return false
		}
		if m == ACCEPTFilterReplay {
			return true
		}
	}
	return true
}

func (b *BaseAppender) SetAppender(ele *LogAppenderXmlEle, writer io.Writer, appender LogAppender) {
	var filters []LogFilter
	var layout *PatternLayout
	if len(ele.Filter) > 0 {
		for _, filter := range ele.Filter {
			if filter.Clazz == "" {
				continue
			}
			nf := GetLogFilterFactory().CreateFilter(filter)
			if nf != nil {
				filters = append(filters, nf)
			}

		}
	}
	layout = NewLayout(ele.Encoder[0].Pattern, writer)

	p := AppenderProperty{
		Filter: filters,
		Layout: layout,
	}

	appender.SetAppenderProperty(&p)
}

type AppenderFactory struct {
	initLock sync.Once
	refMap   map[string]LogAppender
}

func (a *AppenderFactory) init() {
	a.initLock.Do(func() {
		a.refMap = make(map[string]LogAppender)
		a.refMap[ConsoleAppenderAdapterKey] = LogAppender(&ConsoleAppenderImpl{})
		a.refMap[FileAppenderAdapterKey] = LogAppender(&FileAppenderImpl{})
		a.refMap[RollingFileAppenderAdapterKey] = LogAppender(&RollingFileAppenderImpl{})
	})
}

func (a *AppenderFactory) CreateAppender(xml *LogAppenderXmlEle) LogAppender {
	a.init()

	name := strings.ToLower(xml.Clazz)
	appender, ok := a.refMap[name]
	if !ok {
		return nil
	}
	newApp := appender.NewAppender(xml)
	return newApp
}

func (a *AppenderFactory) RegisterAppender(appender LogAppender) {
	a.init()

	key := appender.AppenderKey()
	a.refMap[key] = appender
}

var appenderFactory AppenderFactory = AppenderFactory{}

func GetAppenderFactory() *AppenderFactory {
	return &appenderFactory
}
