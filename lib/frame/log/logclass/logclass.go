package logclass

//Logger, Appenders and Layouts
//the Appender and Layout interfaces are part
// (TRACE, DEBUG, INFO, WARN and ERROR)
// TRACE < DEBUG < INFO <  WARN < ERROR
// an output destination is called an appender
// More than one appender can be attached to a logger
// appender 是往上叠加的

type LogAppender interface {
	Append(row interface{})

	GetLogLayouter() LogLayouter
}

type LogLayouter interface {
}

type LoggerConfig struct {
	Name string
	// 等级 TRACE, DEBUG, INFO, WARN and ERROR
	Level string
	// 是否从上层继承的 默认true
	Extended bool

	// appender 是否往上累加 默认true
	Additivity bool

	Appender []LogAppender

	Parent *LoggerConfig

	Children []*LoggerConfig
}

type LogXmlEle struct {
	Appender []*LogAppenderXmlEle `xml:"appender"`

	Logger []*LogLoggerXmlEle `xml:"logger"`

	Root *LogRootXmlEle `xml:"root"`
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
	Additivity  bool                    `xml:"additivity,attr"`
	AppenderRef []*LogAppenderRefXmlEle `xml:"appender-ref"`
}

type LogRootXmlEle struct {
	Level       string                  `xml:"level,attr"`
	AppenderRef []*LogAppenderRefXmlEle `xml:"appender-ref"`
}
