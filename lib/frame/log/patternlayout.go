package log

import (
	"bytes"
	"context"
	"fmt"
	"goframe/lib/frame/util"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type PatternLayout struct {
	Pattern         string
	HasRuntimeParam bool
	Tpl             *template.Template
	Target          io.Writer
}

//DoLayout err可nil   Target如果是空 就直接返回byte 交给上层处理 ,Target不为空 直接返回空
func (p *PatternLayout) DoLayout(local context.Context, level string, config *LoggerConfig, row string, err interface{}) []byte {
	t1 := time.Now()
	msg := &LogMessage{
		Name:    config.Name,
		Level:   level,
		Msg:     row,
		Err:     err,
		Date:    &t1,
		Context: local,
	}
	if p.HasRuntimeParam {
		_, file2, line, ok := runtime.Caller(3)
		if !ok {
			msg.FileName = "???"
			msg.Line = "0"
		} else {
			msg.FileName = file2
			msg.Line = strconv.Itoa(line)
		}
	}
	if local != nil {
		th := util.ThreadUtil.GetThread(local)
		if th == "" {
			th = "-"
		}
		msg.Thread = th
	}

	// fmt.Println(util.JsonUtil.To2String(msg))
	if err == nil {
		if p.Target != nil {
			err1 := p.Tpl.Execute(p.Target, msg)
			if err1 != nil {
				panic(err1)
			}
			return nil
		} else {
			buf := &bytes.Buffer{}
			err1 := p.Tpl.Execute(buf, msg)
			if err1 != nil {
				panic(err1)
			}
			return buf.Bytes()
		}
	} else {
		rtErr := reflect.ValueOf(err)
		if rtErr.IsZero() {
			if p.Target != nil {
				err1 := p.Tpl.Execute(p.Target, msg)
				if err1 != nil {
					panic(err1)
				}
				return nil
			} else {
				buf := &bytes.Buffer{}
				err1 := p.Tpl.Execute(buf, msg)
				if err1 != nil {
					panic(err1)
				}
				return buf.Bytes()
			}
		} else {
			buf := &bytes.Buffer{}
			err1 := p.Tpl.Execute(buf, msg)
			if err1 != nil {
				panic(err1)
			}

			if rtErr.Type().Implements(util.FrameErrorType) {
				err2 := err.(error)
				buf.Write([]byte(err2.Error()))
			} else {
				err2 := fmt.Sprintf("%s", err)
				buf.Write([]byte(err2))
			}
			stack := strings.Join(strings.Split(string(debug.Stack()), "\n")[7:], "\n")
			buf.Write([]byte(stack))
			buf.Write([]byte("\n"))
			if p.Target != nil {
				p.Target.Write(buf.Bytes())
				return nil
			} else {
				return buf.Bytes()
			}
		}
	}

}

func NewLayout(pattern string, writer io.Writer) *PatternLayout {
	l := &PatternLayout{
		Pattern:         pattern,
		HasRuntimeParam: false,
		Target:          writer,
	}

	f := pattern
	f = date1.ReplaceAllStringFunc(f, func(row string) string {
		p := strings.Index(row, "{")
		dateFormat := "2006-01-02 15:04:05"
		if p >= 0 {
			p1 := strings.Index(row, "}")
			dateFormat = row[p+1 : p1]
		}
		return fmt.Sprintf(`{{logDate "%s" .}}`, dateFormat)
	})

	f = thread1.ReplaceAllStringFunc(f, func(row string) string {
		p := strings.Index(row, "-")
		maxSize := 0
		if p >= 0 {
			p1 := strings.Index(row, "thread")
			maxSizeStr := row[p+1 : p1]
			maxSize, _ = strconv.Atoi(maxSizeStr)
		}
		return fmt.Sprintf(`{{logThread %d .}}`, maxSize)
	})

	f = level1.ReplaceAllStringFunc(f, func(row string) string {
		p := strings.Index(row, "-")
		maxSize := 0
		if p >= 0 {
			p1 := strings.Index(row, "level")
			maxSizeStr := row[p+1 : p1]
			maxSize, _ = strconv.Atoi(maxSizeStr)
		}
		return fmt.Sprintf(`{{logLevel %d .}}`, maxSize)
	})

	//line|file|msg|n|logger

	f = line1.ReplaceAllStringFunc(f, func(row string) string {
		l.HasRuntimeParam = true
		p := strings.Index(row, "-")
		maxSize := 0
		if p >= 0 {
			p1 := strings.Index(row, "line")
			maxSizeStr := row[p+1 : p1]
			maxSize, _ = strconv.Atoi(maxSizeStr)
		}
		return fmt.Sprintf(`{{logLine %d .}}`, maxSize)
	})

	f = file1.ReplaceAllStringFunc(f, func(row string) string {
		l.HasRuntimeParam = true
		p := strings.Index(row, "-")
		maxSize := 0
		if p >= 0 {
			p1 := strings.Index(row, "file")
			maxSizeStr := row[p+1 : p1]
			maxSize, _ = strconv.Atoi(maxSizeStr)
		}
		return fmt.Sprintf(`{{logFile %d .}}`, maxSize)
	})
	//msg|n|logger

	f = msg1.ReplaceAllStringFunc(f, func(row string) string {
		return fmt.Sprintf(`{{logMsg %d .}}`, 0)
	})

	f = br.ReplaceAllStringFunc(f, func(row string) string {
		//return fmt.Sprintf(`{{logBr %d}}`, 0)
		return "\n"
	})
	//%logger{n}

	f = logger1.ReplaceAllStringFunc(f, func(row string) string {
		//return fmt.Sprintf(`{{logBr %d}}`,0)
		p := strings.Index(row, "-")
		maxSize := 0
		if p >= 0 {
			p1 := strings.Index(row, "logger")
			maxSizeStr := row[p+1 : p1]
			maxSize, _ = strconv.Atoi(maxSizeStr)
		}

		p1 := strings.Index(row, "{")
		clazzSize := -1
		if p1 >= 0 {
			p2 := strings.Index(row, "}")
			clazzSizeStr := row[p1+1 : p2]
			clazzSize, _ = strconv.Atoi(clazzSizeStr)
		}
		return fmt.Sprintf(`{{logLogger %d %d .}}`, maxSize, clazzSize)
	})

	f = property1.ReplaceAllStringFunc(f, func(row string) string {
		//return fmt.Sprintf(`{{logBr %d}}`,0)
		p := strings.Index(row, "-")
		maxSize := 0
		if p >= 0 {
			p1 := strings.Index(row, "property")
			maxSizeStr := row[p+1 : p1]
			maxSize, _ = strconv.Atoi(maxSizeStr)
		}

		p1 := strings.Index(row, "{")
		propertyName := "-"
		if p1 >= 0 {
			p2 := strings.Index(row, "}")
			propertyName = row[p1+1 : p2]
		}
		return fmt.Sprintf(`{{propertyLogger %d "%s" .}}`, maxSize, propertyName)
	})

	l.Tpl = template.Must(template.New(fmt.Sprintf("%s-%s-loglayout",
		util.DateUtil.FormatNowByType(util.DatePattern2),
		util.StringUtil.GetRandomStr(7))).Funcs(layOutFuncMap).Parse(f))

	return l
}
