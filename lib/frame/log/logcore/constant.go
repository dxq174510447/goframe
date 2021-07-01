package logcore

const (
	LogAppenderAdapterGroup = "LogAppenderAdapterGroup_"

	ConsoleAppenderAdapterKey = "console"

	FileAppenderAdapterKey = "file"

	TRACELevel = "TRACE"
	DEBUGLevel = "DEBUG"
	INFOLevel  = "INFO"
	WARNLevel  = "WARN"
	ERRORLevel = "ERROR"
)

var LogLevelValue map[string]int = map[string]int{
	TRACELevel: 1,
	DEBUGLevel: 2,
	INFOLevel:  3,
	WARNLevel:  4,
	ERRORLevel: 5,
}
