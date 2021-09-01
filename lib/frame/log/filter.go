package log

import (
	"context"
	"strings"
	"sync"
)

type LevelFilter struct {
	Level      string
	OnMatch    string
	OnMismatch string
}

func (l *LevelFilter) LogDecide(local context.Context, level string, config *LoggerConfig) string {
	if level == l.Level {
		return l.OnMatch
	}
	return l.OnMismatch
}

func (l *LevelFilter) FilterKey() string {
	return LevelFilterAdapterKey
}

func (l *LevelFilter) NewFilter(ele *LogFilterXmlEle) LogFilter {
	onMath := ele.OnMatch
	onMismatch := ele.OnMismatch
	level := ele.Level
	if level == "" {
		level = DEBUGLevel
	}
	if onMath == "" {
		onMath = NEUTRALFilterReplay
	}
	if onMismatch == "" {
		onMismatch = NEUTRALFilterReplay
	}
	return &LevelFilter{
		Level:      strings.ToUpper(level),
		OnMatch:    strings.ToUpper(onMath),
		OnMismatch: strings.ToUpper(onMismatch),
	}
}

type ThresholdFilter struct {
	Level string
}

func (l *ThresholdFilter) LogDecide(local context.Context, level string, config *LoggerConfig) string {
	m := LogLevelValue[level] - LogLevelValue[l.Level]
	if m >= 0 {
		return NEUTRALFilterReplay
	}
	return DENYFilterReplay
}

func (l *ThresholdFilter) FilterKey() string {
	return ThresholdFilterAdapterKey
}
func (l *ThresholdFilter) NewFilter(ele *LogFilterXmlEle) LogFilter {
	level := ele.Level
	if level == "" {
		level = DEBUGLevel
	}
	return &ThresholdFilter{
		Level: strings.ToUpper(level),
	}
}

type LogFilterFactory struct {
	initLock sync.Once
	refMap   map[string]LogFilter
}

func (a *LogFilterFactory) init() {
	a.initLock.Do(func() {
		a.refMap = make(map[string]LogFilter)
		a.refMap[LevelFilterAdapterKey] = LogFilter(&LevelFilter{})
		a.refMap[ThresholdFilterAdapterKey] = LogFilter(&ThresholdFilter{})
	})
}

func (a *LogFilterFactory) CreateFilter(xml *LogFilterXmlEle) LogFilter {
	a.init()

	name := strings.ToLower(xml.Clazz)
	filter, ok := a.refMap[name]
	if !ok {
		return nil
	}
	newFilter := filter.NewFilter(xml)
	return newFilter
}

func (a *LogFilterFactory) RegisterFilter(filter LogFilter) {
	a.init()

	key := filter.FilterKey()
	a.refMap[key] = filter
}

var logFilterFactory LogFilterFactory = LogFilterFactory{}

func GetLogFilterFactory() *LogFilterFactory {
	return &logFilterFactory
}
