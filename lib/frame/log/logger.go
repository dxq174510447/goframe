package log

import (
	"context"
	"fmt"
)

type Logger struct {
	Config *LoggerConfig
}

func (l *Logger) Trace(local context.Context, format string, a ...interface{}) {
	if !l.IsTraceEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(TRACELevel, current) {
		for _, appender := range current.Appender {
			appender.AppendRow(local, TRACELevel, l.Config, content, nil)
		}
		if !current.Additivity {
			break
		}
		current = current.Parent
	}
}

func (l *Logger) IsTraceEnable() bool {
	return l.isLevelEnable(TRACELevel, l.Config)
}

func (l *Logger) Debug(local context.Context, format string, a ...interface{}) {
	if !l.IsDebugEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(DEBUGLevel, current) {
		for _, appender := range current.Appender {
			appender.AppendRow(local, DEBUGLevel, l.Config, content, nil)
		}
		if !current.Additivity {
			break
		}
		current = current.Parent
	}
}

func (l *Logger) isLevelEnable(currentLevel string, targetLevel *LoggerConfig) bool {
	return LogLevelValue[currentLevel] >= LogLevelValue[targetLevel.Level]
}

func (l *Logger) IsDebugEnable() bool {
	return l.isLevelEnable(DEBUGLevel, l.Config)
}

func (l *Logger) Info(local context.Context, format string, a ...interface{}) {
	if !l.IsInfoEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(INFOLevel, current) {
		for _, appender := range current.Appender {
			appender.AppendRow(local, INFOLevel, l.Config, content, nil)
		}
		if !current.Additivity {
			break
		}
		current = current.Parent
	}
}

func (l *Logger) IsInfoEnable() bool {
	return l.isLevelEnable(INFOLevel, l.Config)
}

func (l *Logger) Warn(local context.Context, format string, a ...interface{}) {
	if !l.IsWarnEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(WARNLevel, current) {
		for _, appender := range current.Appender {
			appender.AppendRow(local, WARNLevel, l.Config, content, nil)
		}
		if !current.Additivity {
			break
		}
		current = current.Parent
	}
}

func (l *Logger) IsWarnEnable() bool {
	return l.isLevelEnable(WARNLevel, l.Config)
}

func (l *Logger) Error(local context.Context, err interface{}, format string, a ...interface{}) {
	if !l.IsErrorEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(ERRORLevel, current) {
		for _, appender := range current.Appender {
			appender.AppendRow(local, ERRORLevel, l.Config, content, err)
		}
		if !current.Additivity {
			break
		}
		current = current.Parent
	}
}

func (l *Logger) IsErrorEnable() bool {
	return l.isLevelEnable(ERRORLevel, l.Config)
}
