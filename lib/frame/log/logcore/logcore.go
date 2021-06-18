package logcore

import (
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

// %date %date{HH:mm:ss.SSS} %{-n}thread %{-n}level %logger{n} %line %file %msg %n
type LogMessage struct {
	Name     string
	Level    string
	Line     string
	FileName string
	Msg      string
	Err      interface{}
	Thread   string
	Date     time.Time
}
type PatternLayout struct {
	Pattern         string
	HasRuntimeParam bool
	Tpl             *template.Template
	Target          io.Writer
}

// err可nil
func (p *PatternLayout) DoLayout(local *context.LocalStack, config *logclass.LoggerConfig, row string, err interface{}) {
	msg := &LogMessage{
		Name:  config.Name,
		Level: config.Level,
		Msg:   row,
		Err:   err,
		Date:  time.Now(),
	}
	if p.HasRuntimeParam {
		_, file, line, ok := runtime.Caller(3)
		if !ok {
			msg.FileName = "???"
			msg.Line = "0"
		} else {
			msg.FileName = util.ClassUtil.GetJavaFileNameByType(file)
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
	err1 := p.Tpl.Execute(p.Target, msg)
	if err1 != nil {
		panic(err1)
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

	l.Tpl = template.Must(template.New(fmt.Sprintf("%s-%s-loglayout",
		util.DateUtil.FormatNowByType(util.DatePattern2),
		util.StringUtil.GetRandomStr(7))).Funcs(layOutFuncMap).Parse(f))

	return l
}

type ConsoleAppenderImpl struct {
	Layout *PatternLayout
	Target io.Writer
}

func (c *ConsoleAppenderImpl) AppenderKey() string {
	return ConsoleAppenderAdapterKey
}

func (c *ConsoleAppenderImpl) NewAppender(ele *logclass.LogAppenderXmlEle) logclass.LogAppender {
	layout := NewLayout(ele.Encoder[0].Pattern, os.Stdout)
	result := &ConsoleAppenderImpl{
		Layout: layout,
		Target: os.Stdout,
	}
	return logclass.LogAppender(result)
}

func (c *ConsoleAppenderImpl) AppendRow(local *context.LocalStack, config *logclass.LoggerConfig, row string, err interface{}) {
	c.Layout.DoLayout(local, config, row, err)
	os.Stdout.Sync()
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
			appender.AppendRow(local, l.Config, content, nil)
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
			appender.AppendRow(local, l.Config, content, nil)
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
			appender.AppendRow(local, l.Config, content, nil)
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
			appender.AppendRow(local, l.Config, content, nil)
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
			appender.AppendRow(local, l.Config, content, err)
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

var buildInAppender map[string]logclass.LogAppender = map[string]logclass.LogAppender{
	ConsoleAppenderAdapterKey: logclass.LogAppender(&console),
}

type LogFactory struct {
	Appender map[string]logclass.LogAppender
	Root     *logclass.LoggerConfig
	RefMap   map[string]*logclass.LoggerConfig
}

func (l *LogFactory) GetLoggerType(p reflect.Type) application.AppLoger {
	name := util.ClassUtil.GetJavaClassNameByType(p)
	return l.GetLoggerString(name)
}

func (l *LogFactory) GetLoggerString(name string) application.AppLoger {
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

var logFactory LogFactory = LogFactory{}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&logFactory))
}
