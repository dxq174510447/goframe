package log

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"goframe/lib/frame/application"
	"goframe/lib/frame/util"
	"reflect"
	"strings"
	"sync"
	"text/template"
)

var defaultLogback = `
<configuration> 
  
  <appender name="DEFAULT_CONSOLE" class="console">
    <encoder>
      <pattern>%date [%-5thread] [%-5level] %-10logger{0}.%-5line %msg %n</pattern>
    </encoder>
  </appender>

  <root level="debug">
	<appender-ref ref="DEFAULT_CONSOLE"/>
  </root>
</configuration>
`

type LoggerFactory struct {
	Appender map[string]LogAppender
	Root     *LoggerConfig
	RefMap   map[string]*LoggerConfig
	initLock sync.Once
}

func (l *LoggerFactory) init() {
	l.initLock.Do(func() {
		l.RefMap = make(map[string]*LoggerConfig)
		l.Appender = make(map[string]LogAppender)
		l.Parse(defaultLogback, nil)
	})
}

func (l *LoggerFactory) GetLoggerType(p reflect.Type) application.AppLoger {
	name := util.ClassUtil.GetClassNameByTypeV1(p)
	return l.GetLoggerString(name)
}

func (l *LoggerFactory) LogTypeName() string {
	return PlatLogKey
}

func (l *LoggerFactory) GetLoggerString(name string) application.AppLoger {
	var node *LoggerConfig
	if config, ok := l.RefMap[name]; ok {
		node = config
	} else {
		node1 := &LoggerConfig{
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

func (l *LoggerFactory) innerPrintTree(node *LoggerConfig, depth int) {
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
func (l *LoggerFactory) PrintTree() {
	fmt.Println(l.Root.Level, len(l.Root.Appender))
	for _, c := range l.Root.Children {
		l.innerPrintTree(c, 0)
	}
}

func (l *LoggerFactory) Cover2Logger(ele *LogLoggerXmlEle) *LoggerConfig {
	config := &LoggerConfig{
		Name:        ele.Name,
		Level:       ele.Level,
		Additivity:  true,
		ChildrenMap: make(map[string]*LoggerConfig),
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

func (l *LoggerFactory) AddLevelNode(node *LoggerConfig) {
	if node.Name == "" {
		return
	}
	current := l.Root
	keys := strings.Split(node.Name, "/")
	lsize := len(keys)
	if l.RefMap == nil {
		l.RefMap = make(map[string]*LoggerConfig)
	}
	for i := 0; i < lsize; i++ {
		key := keys[i]
		if children, ok := current.ChildrenMap[key]; ok {
			if i == (lsize - 1) {

				if node.Level != "" && node.Extended == 1 {
					node.Extended = 0
				}

				children.Level = node.Level
				children.Additivity = node.Additivity
				children.Appender = node.Appender
				children.Level = strings.ToUpper(children.Level)
			}
			current = children
		} else {
			c := &LoggerConfig{
				Name:        strings.Join(keys[0:i+1], "/"),
				Level:       "",
				Additivity:  true,
				Parent:      current,
				ChildrenMap: make(map[string]*LoggerConfig),
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

func (l *LoggerFactory) ParseAndReload(content string, funcMap template.FuncMap) {
	l.Parse(content, funcMap)
	l.ReloadLevel()
}

// ReloadLevel 重建level等级 有些level是继承的
func (l *LoggerFactory) ReloadLevel() {
	if l.Root == nil {
		return
	}
	for _, v := range l.Root.ChildrenMap {
		l.reloadLeveling(v)
	}
}

func (l *LoggerFactory) reloadLeveling(node *LoggerConfig) {

	if node.Level != "" && node.Extended == 1 {
		current := node.Parent
		for current != nil {
			if current.Level != "" {
				break
			}
			current = current.Parent
		}
		node.Level = current.Level
	}

	if len(node.ChildrenMap) == 0 {
		return
	}
	for _, v := range node.ChildrenMap {
		l.reloadLeveling(v)
	}

}

func (l *LoggerFactory) Parse(content string, funcMap template.FuncMap) {
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

	config := &LogXmlEle{}

	err1 := xml.Unmarshal([]byte(xml1), config)

	if err1 != nil {
		panic(err1)
	}

	if l.Appender == nil {
		l.Appender = make(map[string]LogAppender)
	}

	for _, xml := range config.Appender {
		newApp := GetAppenderFactory().CreateAppender(xml)
		if newApp == nil {
			continue
		}
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
			root = &LogLoggerXmlEle{
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

var loggerFactory LoggerFactory = LoggerFactory{}

func GetLoggerFactory() *LoggerFactory {
	loggerFactory.init()
	return &loggerFactory
}

func init() {
	loggerFactory.init()
	application.GetResourcePool().RegisterLogFactory(application.AppLogFactoryer(&loggerFactory))
}
