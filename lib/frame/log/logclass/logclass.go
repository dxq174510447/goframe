package logclass

//Logger, Appenders and Layouts
//the Appender and Layout interfaces are part
// (TRACE, DEBUG, INFO, WARN and ERROR)
// TRACE < DEBUG < INFO <  WARN < ERROR
// an output destination is called an appender
// More than one appender can be attached to a logger
// appender 是往上叠加的

type LogAppender interface {
	AppendRow(row string)
	AppenderKey() string
	NewAppender(ele *LogAppenderXmlEle) LogAppender
}

type LogLayouter interface {
}

type LoggerConfig struct {
	Name string
	// 等级 TRACE, DEBUG, INFO, WARN and ERROR
	// 设置的时候 转大写
	Level string

	// level是否是继承的 0不是继承 1是继承
	Extended int

	// appender 是否往上累加 默认true
	Additivity bool

	Appender []LogAppender

	Parent *LoggerConfig

	Children []*LoggerConfig

	ChildrenMap map[string]*LoggerConfig
}

type LogXmlEle struct {
	Appender []*LogAppenderXmlEle `xml:"appender"`

	Logger []*LogLoggerXmlEle `xml:"logger"`

	Root *LogLoggerXmlEle `xml:"root"`
}

type LogAppenderEncodeXmlEle struct {
	Pattern string `xml:"pattern"`
}
type LogAppenderXmlEle struct {
	Name    string                     `xml:"name,attr"`
	Clazz   string                     `xml:"class,attr"`
	Encoder []*LogAppenderEncodeXmlEle `xml:"encoder"`
}

type LogAppenderRefXmlEle struct {
	Ref string `xml:"ref,attr"`
}

type LogLoggerXmlEle struct {
	Name        string                  `xml:"name,attr"`
	Level       string                  `xml:"level,attr"`
	Additivity  string                  `xml:"additivity,attr"`
	AppenderRef []*LogAppenderRefXmlEle `xml:"appender-ref"`
}
