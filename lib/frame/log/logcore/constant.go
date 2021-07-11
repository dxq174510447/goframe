package logcore

import (
	"strings"
	"time"
)

const (
	LogAppenderAdapterGroup = "LogAppenderAdapterGroup_"

	ConsoleAppenderAdapterKey     = "console"
	FileAppenderAdapterKey        = "file"
	RollingFileAppenderAdapterKey = "rolling_file"

	TRACELevel = "TRACE"
	DEBUGLevel = "DEBUG"
	INFOLevel  = "INFO"
	WARNLevel  = "WARN"
	ERRORLevel = "ERROR"

	LevelFilterAdapterKey     = "level"
	ThresholdFilterAdapterKey = "threshold"

	DENYFilterReplay    = "DENY"
	NEUTRALFilterReplay = "NEUTRAL"
	ACCEPTFilterReplay  = "ACCEPT"

	//rolling_rule
	TOP_OF_MINUTE = "TOP_OF_MINUTE_"
	TOP_OF_HOUR   = "TOP_OF_HOUR_"
	//HALF_DAY = "HALF_DAY_"
	TOP_OF_DAY = "TOP_OF_DAY_"
	//TOP_OF_WEEK = "TOP_OF_WEEK_"
	TOP_OF_MONTH = "TOP_OF_MONTH_"
)

var LogRollingType []string = []string{TOP_OF_MINUTE, TOP_OF_HOUR, TOP_OF_DAY, TOP_OF_MONTH}

var LogLevelValue map[string]int = map[string]int{
	TRACELevel: 1,
	DEBUGLevel: 2,
	INFOLevel:  3,
	WARNLevel:  4,
	ERRORLevel: 5,
}

func GetTimePatternFromFileNamePattern(fileNamePattern string) string {
	p := strings.LastIndex(fileNamePattern, "%d")
	if p == -1 {
		panic("rolling filename error")
	}
	begin := p + 2
	if begin == len(fileNamePattern) || fileNamePattern[begin:begin+1] != "{" {
		return "2006-01-02"
	} else {
		last := strings.LastIndex(fileNamePattern, "}")
		return fileNamePattern[begin+1 : last]
	}
}

func GetRollingRule(rollingRule string, time2 *time.Time, period int) *time.Time {
	switch rollingRule {
	case TOP_OF_MINUTE:
		t := time2.Add(time.Minute * time.Duration(period))
		return &t
	case TOP_OF_HOUR:
		t := time2.Add(time.Hour * time.Duration(period))
		return &t
	case TOP_OF_DAY:
		t := time2.AddDate(0, 0, period)
		return &t
	case TOP_OF_MONTH:
		t := time2.AddDate(0, period, 0)
		return &t
	}
	return nil
}
