package log

import (
	"fmt"
	"goframe/lib/frame/util"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	PlatLogKey = "PlatLogKey_"

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

var date1 = regexp.MustCompile("%date(\\{[^\\}]+\\})?")
var thread1 = regexp.MustCompile("%(\\-\\d+)?thread")
var level1 = regexp.MustCompile("%(\\-\\d+)?level")
var line1 = regexp.MustCompile("%(\\-\\d+)?line")
var file1 = regexp.MustCompile("%(\\-\\d+)?file")
var msg1 = regexp.MustCompile("%(\\-\\d+)?msg")
var br = regexp.MustCompile("%(\\-\\d+)?n")
var logger1 = regexp.MustCompile("%(\\-\\d+)?logger(\\{[^\\}]+\\})?")

var property1 = regexp.MustCompile("%(\\-\\d+)?property(\\{[^\\}]+\\})?")

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

var layOutFuncMap = template.FuncMap{
	"logDate": func(format string, msg *LogMessage) string {
		return util.DateUtil.FormatByType(msg.Date, format)
	},
	"logThread": func(size int, msg *LogMessage) string {
		if len(msg.Thread) >= size {
			return msg.Thread
		}

		n := []byte(msg.Thread)
		sp := make([]byte, size, size)
		for i := 0; i < size; i++ {
			sp[i] = 32
		}
		copy(sp, n)
		return string(sp)
	},
	"logLevel": func(size int, msg *LogMessage) string {
		if len(msg.Level) >= size {
			return msg.Level
		}

		n := []byte(msg.Level)
		sp := make([]byte, size, size)
		for i := 0; i < size; i++ {
			sp[i] = 32
		}
		copy(sp, n)
		return string(sp)
	},
	"logLine": func(size int, msg *LogMessage) string {
		if len(msg.Line) >= size {
			return msg.Line
		}

		n := []byte(msg.Line)
		sp := make([]byte, size, size)
		for i := 0; i < size; i++ {
			sp[i] = 32
		}
		copy(sp, n)
		return string(sp)
	},
	"logFile": func(size int, msg *LogMessage) string {
		if len(msg.FileName) >= size {
			return msg.FileName
		}

		n := []byte(msg.FileName)
		sp := make([]byte, size, size)
		for i := 0; i < size; i++ {
			sp[i] = 32
		}
		copy(sp, n)
		return string(sp)
	},
	"logMsg": func(size int, msg *LogMessage) string {
		return msg.Msg
	},
	"logBr": func(size int, msg *LogMessage) string {
		return "\n"
	},
	"logLogger": func(size int, clazzSize int, msg *LogMessage) string {
		className := msg.Name
		if clazzSize == 0 {
			p := strings.LastIndex(className, ".")
			className = className[p+1:]
		}

		if len(className) >= size {
			return className
		}

		n := []byte(className)
		sp := make([]byte, size, size)
		for i := 0; i < size; i++ {
			sp[i] = 32
		}
		copy(sp, n)
		return string(sp)
	},
	"propertyLogger": func(size int, propertyName string, msg *LogMessage) string {
		var rowMsg string = "-"
		if msg.Context != nil {
			r := msg.Context.Value(propertyName)
			if r != nil {
				rv := reflect.ValueOf(r)
				switch rv.Kind() {
				case reflect.String:
					rowMsg = r.(string)
				case reflect.Int:
					r1 := r.(int)
					rowMsg = strconv.Itoa(r1)
				case reflect.Int64:
					r1 := r.(int64)
					rowMsg = strconv.FormatInt(r1, 10)
				default:
					rowMsg = fmt.Sprintf("%s格式识别不到", propertyName)
				}
			}
		}
		return rowMsg
	},
}
