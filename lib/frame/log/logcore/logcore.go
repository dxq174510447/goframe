package logcore

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"io"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Release lock while getting caller info - it's expensive.
//l.mu.Unlock()
//var ok bool
//_, file, line, ok = runtime.Caller(calldepth)
//if !ok {
//file = "???"
//  line = 0
//}
//l.mu.Lock()
var date1 = regexp.MustCompile("%date(\\{[^\\}]+\\})?")
var thread1 = regexp.MustCompile("%(\\-\\d+)?thread")
var level1 = regexp.MustCompile("%(\\-\\d+)?level")
var line1 = regexp.MustCompile("%(\\-\\d+)?line")
var file1 = regexp.MustCompile("%(\\-\\d+)?file")
var msg1 = regexp.MustCompile("%(\\-\\d+)?msg")
var br = regexp.MustCompile("%(\\-\\d+)?n")
var logger1 = regexp.MustCompile("%(\\-\\d+)?logger(\\{[^\\}]+\\})?")

var property1 = regexp.MustCompile("%(\\-\\d+)?property(\\{[^\\}]+\\})?")

// %date %date{HH:mm:ss.SSS} %{-n}thread %{-n}level %logger{n} %line %file %msg %n
type LogMessage struct {
	Name     string
	Level    string
	Line     string
	FileName string
	Msg      string
	Err      interface{}
	Thread   string
	Date     *time.Time
	Context  *context.LocalStack
}

type PatternLayout struct {
	Pattern         string
	HasRuntimeParam bool
	Tpl             *template.Template
	Target          io.Writer
}

//DoLayout err可nil   Target如果是空 就直接返回byte 交给上层处理 ,Target不为空 直接返回空
func (p *PatternLayout) DoLayout(local *context.LocalStack, level string, config *logclass.LoggerConfig, row string, err interface{}) []byte {
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
			msg.FileName = util.ClassUtil.GetJavaFileNameByType(file2)
			msg.Line = strconv.Itoa(line)
		}
	}
	if local != nil {
		th := local.GetThread()
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
			r := msg.Context.Get(propertyName)
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

// SetAppender 设置appender一些公用节点
func SetAppender(ele *logclass.LogAppenderXmlEle, writer io.Writer, appender logclass.LogAppender) {
	var filters []logclass.LogFilter
	var layout *PatternLayout
	if len(ele.Filter) > 0 {
		for _, filter := range ele.Filter {
			if filter.Clazz == "" {
				continue
			}
			if f, ok := buildInFilter[strings.ToLower(filter.Clazz)]; ok {
				nf := f.NewFilter(filter)
				if nf != nil {
					filters = append(filters, nf)
				}
			}
		}
	}
	layout = newLayout(ele.Encoder[0].Pattern, writer)

	p := logclass.AppenderProperty{
		Filter: filters,
		Layout: layout,
	}

	appender.SetAppenderProperty(&p)
}

func newLayout(pattern string, writer io.Writer) *PatternLayout {
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

type FileAppenderImpl struct {
	Property   *logclass.AppenderProperty
	Target     io.Writer
	FilePath   string
	FileTarget *os.File
	FileBuffer *bufio.Writer
}

func (f *FileAppenderImpl) AppendRow(local *context.LocalStack, level string, config *logclass.LoggerConfig, row string, err interface{}) {
	if IsAppendRow(local, level, config, f.Property) {
		f.Property.Layout.DoLayout(local, level, config, row, err)
		f.FileBuffer.Flush()
	}
}

func (f *FileAppenderImpl) AppenderKey() string {
	return FileAppenderAdapterKey
}

func (f *FileAppenderImpl) SetAppenderProperty(property *logclass.AppenderProperty) {
	f.Property = property
}

func (f *FileAppenderImpl) NewAppender(ele *logclass.LogAppenderXmlEle) logclass.LogAppender {

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
	SetAppender(ele, bufferWriter, result)
	return logclass.LogAppender(result)
}

type ConsoleAppenderImpl struct {
	Property *logclass.AppenderProperty
	Target   io.Writer
}

func (c *ConsoleAppenderImpl) AppenderKey() string {
	return ConsoleAppenderAdapterKey
}

func (c *ConsoleAppenderImpl) NewAppender(ele *logclass.LogAppenderXmlEle) logclass.LogAppender {
	//layout := NewLayout(ele.Encoder[0].Pattern, os.Stdout)
	result := &ConsoleAppenderImpl{
		Target: os.Stdout,
	}
	SetAppender(ele, os.Stdout, result)
	return logclass.LogAppender(result)
}

func (c *ConsoleAppenderImpl) SetAppenderProperty(property *logclass.AppenderProperty) {
	c.Property = property
}

func (c *ConsoleAppenderImpl) AppendRow(local *context.LocalStack, level string, config *logclass.LoggerConfig, row string, err interface{}) {
	if IsAppendRow(local, level, config, c.Property) {
		c.Property.Layout.DoLayout(local, level, config, row, err)
		os.Stdout.Sync()
	}
}

type LevelFilter struct {
	Level      string
	OnMatch    string
	OnMismatch string
}

func (l *LevelFilter) LogDecide(local *context.LocalStack, level string, config *logclass.LoggerConfig) string {
	if level == l.Level {
		return l.OnMatch
	}
	return l.OnMismatch
}

func (l *LevelFilter) FilterKey() string {
	return LevelFilterAdapterKey
}

func (l *LevelFilter) NewFilter(ele *logclass.LogFilterXmlEle) logclass.LogFilter {
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

func (l *ThresholdFilter) LogDecide(local *context.LocalStack, level string, config *logclass.LoggerConfig) string {
	m := LogLevelValue[level] - LogLevelValue[l.Level]
	if m >= 0 {
		return NEUTRALFilterReplay
	}
	return DENYFilterReplay
}

func (l *ThresholdFilter) FilterKey() string {
	return ThresholdFilterAdapterKey
}
func (l *ThresholdFilter) NewFilter(ele *logclass.LogFilterXmlEle) logclass.LogFilter {
	level := ele.Level
	if level == "" {
		level = DEBUGLevel
	}
	return &ThresholdFilter{
		Level: strings.ToUpper(level),
	}
}

type Logger struct {
	Config *logclass.LoggerConfig
}

func (l *Logger) Trace(local *context.LocalStack, format string, a ...interface{}) {
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

func (l *Logger) Debug(local *context.LocalStack, format string, a ...interface{}) {
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

func (l *Logger) isLevelEnable(currentLevel string, targetLevel *logclass.LoggerConfig) bool {
	return LogLevelValue[currentLevel] >= LogLevelValue[targetLevel.Level]
}

func (l *Logger) IsDebugEnable() bool {
	return l.isLevelEnable(DEBUGLevel, l.Config)
}

func (l *Logger) Info(local *context.LocalStack, format string, a ...interface{}) {
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

func (l *Logger) Warn(local *context.LocalStack, format string, a ...interface{}) {
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

func (l *Logger) Error(local *context.LocalStack, err interface{}, format string, a ...interface{}) {
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

var console ConsoleAppenderImpl = ConsoleAppenderImpl{}
var file FileAppenderImpl = FileAppenderImpl{}
var rollingfile RollingFileAppenderImpl = RollingFileAppenderImpl{}

var buildInAppender map[string]logclass.LogAppender = map[string]logclass.LogAppender{
	ConsoleAppenderAdapterKey:     logclass.LogAppender(&console),
	FileAppenderAdapterKey:        logclass.LogAppender(&file),
	RollingFileAppenderAdapterKey: logclass.LogAppender(&rollingfile),
}

var levelFilter LevelFilter = LevelFilter{}
var thresholdFilter ThresholdFilter = ThresholdFilter{}

var buildInFilter map[string]logclass.LogFilter = map[string]logclass.LogFilter{
	LevelFilterAdapterKey:     logclass.LogFilter(&levelFilter),
	ThresholdFilterAdapterKey: logclass.LogFilter(&thresholdFilter),
}

type LogFactory struct {
	Appender map[string]logclass.LogAppender
	Root     *logclass.LoggerConfig
	RefMap   map[string]*logclass.LoggerConfig
}

func (l *LogFactory) GetLoggerType(p reflect.Type) logclass.AppLoger {
	name := util.ClassUtil.GetJavaClassNameByType(p)
	return l.GetLoggerString(name)
}

func (l *LogFactory) GetLoggerString(name string) logclass.AppLoger {
	var node *logclass.LoggerConfig
	if config, ok := l.RefMap[name]; ok {
		node = config
	} else {
		node1 := &logclass.LoggerConfig{
			Name:       name,
			Level:      "",
			Additivity: true,
		}
		l.AddLevelNode(node1)
		node = l.RefMap[name]
	}
	//l.PrintTree()
	//fmt.Println(node.Name)
	if node.Level == "" {
		current := node.Parent
		for current != nil {
			if current.Level != "" {
				break
			}
			current = current.Parent
		}
		node.Extended = 1
		node.Level = current.Level
	}
	return &Logger{
		Config: node,
	}
}

func (l *LogFactory) innerPrintTree(node *logclass.LoggerConfig, depth int) {
	m := make([]string, depth, depth)
	for i := 0; i < depth; i++ {
		m[i] = "-"
	}

	fmt.Println(strings.Join(m, ""), node.Name, node.Level, len(node.Appender))
	if len(node.Children) > 0 {
		for _, c := range node.Children {
			l.innerPrintTree(c, depth+1)
		}
	}
}
func (l *LogFactory) PrintTree() {
	fmt.Println(l.Root.Level, len(l.Root.Appender))
	for _, c := range l.Root.Children {
		l.innerPrintTree(c, 0)
	}
}

func (l *LogFactory) Cover2Logger(ele *logclass.LogLoggerXmlEle) *logclass.LoggerConfig {
	config := &logclass.LoggerConfig{
		Name:        ele.Name,
		Level:       ele.Level,
		Additivity:  true,
		ChildrenMap: make(map[string]*logclass.LoggerConfig),
	}
	if ele.Additivity != "" && strings.ToLower(ele.Additivity) == "false" {
		config.Additivity = false
	}
	for _, appender := range ele.AppenderRef {
		key := appender.Ref
		if app, ok := l.Appender[key]; ok {
			config.Appender = append(config.Appender, app)
		}
	}
	return config
}

func (l *LogFactory) AddLevelNode(node *logclass.LoggerConfig) {
	if node.Name == "" {
		return
	}
	current := l.Root
	keys := strings.Split(node.Name, ".")
	lsize := len(keys)
	if l.RefMap == nil {
		l.RefMap = make(map[string]*logclass.LoggerConfig)
	}
	for i := 0; i < lsize; i++ {
		key := keys[i]
		if children, ok := current.ChildrenMap[key]; ok {
			if i == (lsize - 1) {
				children.Level = node.Level
				children.Additivity = node.Additivity
				children.Appender = node.Appender
				children.Level = strings.ToUpper(children.Level)
			}
			current = children
		} else {
			c := &logclass.LoggerConfig{
				Name:        strings.Join(keys[0:i+1], "."),
				Level:       "",
				Additivity:  true,
				Parent:      current,
				ChildrenMap: make(map[string]*logclass.LoggerConfig),
			}
			if i == (lsize - 1) {
				// last
				c.Level = node.Level
				c.Additivity = node.Additivity
				c.Appender = node.Appender
			}
			c.Level = strings.ToUpper(c.Level)
			current.ChildrenMap[key] = c
			current.Children = append(current.Children, c)
			l.RefMap[c.Name] = c
			current = c
		}
	}
}
func (l *LogFactory) Parse(content string, funcMap template.FuncMap) {

	var tpl *template.Template
	if funcMap == nil || len(funcMap) == 0 {
		tpl = template.Must(template.New(fmt.Sprintf("%s-%s-logcore", util.DateUtil.FormatNowByType(util.DatePattern2), util.StringUtil.GetRandomStr(5))).Parse(content))
	} else {
		tpl = template.Must(template.New(fmt.Sprintf("%s-%s-logcore", util.DateUtil.FormatNowByType(util.DatePattern2), util.StringUtil.GetRandomStr(5))).Funcs(funcMap).Parse(content))
	}

	buf := &bytes.Buffer{}
	param := make(map[string]interface{})
	err := tpl.Execute(buf, param)
	if err != nil {
		panic(err)
	}
	xml1 := util.StringUtil.RemoveEmptyRow(buf.String())

	config := &logclass.LogXmlEle{}

	err1 := xml.Unmarshal([]byte(xml1), config)

	if err1 != nil {
		panic(err1)
	}

	if l.Appender == nil {
		l.Appender = make(map[string]logclass.LogAppender)
	}

	for _, xml := range config.Appender {
		name := strings.ToLower(xml.Clazz)
		appender, ok := buildInAppender[name]
		if !ok {
			continue
		}
		newApp := appender.NewAppender(xml)
		l.Appender[xml.Name] = newApp
	}

	if l.Root != nil {
		root := config.Root
		if root != nil {
			treeRoot := l.Cover2Logger(root)
			//l.Root.Name = treeRoot.Name
			l.Root.Level = treeRoot.Level
			l.Root.Additivity = treeRoot.Additivity
			l.Root.Appender = treeRoot.Appender
			if l.Root.Level == "" {
				l.Root.Level = DEBUGLevel
			}
			l.Root.Level = strings.ToUpper(l.Root.Level)
		}
	} else {
		root := config.Root
		if root == nil {
			root = &logclass.LogLoggerXmlEle{
				Level: DEBUGLevel,
				Name:  "root",
			}
		} else {
			root.Name = "root"
		}
		if root.Level == "" {
			root.Level = DEBUGLevel
		}
		treeRoot := l.Cover2Logger(root)
		l.Root = treeRoot
		l.Root.Level = strings.ToUpper(l.Root.Level)
	}

	for _, xml := range config.Logger {
		node := l.Cover2Logger(xml)
		l.AddLevelNode(node)
	}
	//for k, _ := range l.RefMap {
	//	fmt.Println(k)
	//}
	//l.PrintTree()
}

func (l *LogFactory) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

func AddLogAppender(appender logclass.LogAppender) {
	key := appender.AppenderKey()
	buildInAppender[key] = appender
}

func AddLogFilter(filter logclass.LogFilter) {
	key := filter.FilterKey()
	buildInFilter[key] = filter
}

//IsAppendRow 是否过滤掉日志
func IsAppendRow(local *context.LocalStack, level string, config *logclass.LoggerConfig, property *logclass.AppenderProperty) bool {
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

var logFactory LogFactory = LogFactory{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&logFactory))
}
