package logcore

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type ConsoleAppenderImpl struct {
	Pattern string
}

func (c *ConsoleAppenderImpl) AppenderKey() string {
	return ConsoleAppenderAdapterKey
}

func (c *ConsoleAppenderImpl) NewAppender(ele *logclass.LogAppenderXmlEle) logclass.LogAppender {
	result := &ConsoleAppenderImpl{
		Pattern: ele.Encoder[0].Pattern,
	}
	return logclass.LogAppender(result)
}

func (c *ConsoleAppenderImpl) AppendRow(row string) {
	os.Stdout.WriteString(row + "\n")
}

type Logger struct {
	Config *logclass.LoggerConfig
}

func (l *Logger) Trace(format string, a ...interface{}) {
	if !l.IsTraceEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(TRACELevel, current) {
		for _, appender := range l.Config.Appender {
			appender.AppendRow(content)
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

func (l *Logger) Debug(format string, a ...interface{}) {
	if !l.IsDebugEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(DEBUGLevel, current) {
		for _, appender := range l.Config.Appender {
			appender.AppendRow(content)
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

func (l *Logger) Info(format string, a ...interface{}) {
	if !l.IsInfoEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(INFOLevel, current) {
		for _, appender := range l.Config.Appender {
			appender.AppendRow(content)
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

func (l *Logger) Warn(format string, a ...interface{}) {
	if !l.IsWarnEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(WARNLevel, current) {
		for _, appender := range l.Config.Appender {
			appender.AppendRow(content)
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

func (l *Logger) Error(err error, format string, a ...interface{}) {
	if !l.IsErrorEnable() {
		return
	}

	content := fmt.Sprintf(format, a...)
	current := l.Config
	for current != nil && l.isLevelEnable(ERRORLevel, current) {
		for _, appender := range l.Config.Appender {
			appender.AppendRow(content)
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
	fmt.Println(name)
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
	if node.Level == "" {
		current := node.Parent
		for current != nil {
			if current.Level != "" {
				break
			}
			current = node.Parent
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
		tpl = template.Must(template.New(fmt.Sprintf("%s-logcore", util.DateUtil.FormatNowByType(util.DatePattern2))).Parse(content))
	} else {
		tpl = template.Must(template.New(fmt.Sprintf("%s-logcore", util.DateUtil.FormatNowByType(util.DatePattern2))).Funcs(funcMap).Parse(content))
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

func AddLogAppender(appender logclass.LogAppender) {
	key := appender.AppenderKey()
	buildInAppender[key] = appender
}

var logFactory LogFactory = LogFactory{}

func GetLogFactory() *LogFactory {
	return &logFactory
}
