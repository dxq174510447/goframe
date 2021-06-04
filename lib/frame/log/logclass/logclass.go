package logclass

//Logger, Appenders and Layouts
//the Appender and Layout interfaces are part
// (TRACE, DEBUG, INFO, WARN and ERROR)
// TRACE < DEBUG < INFO <  WARN < ERROR
// an output destination is called an appender
// More than one appender can be attached to a logger
// appender 是往上叠加的

type LogAppender struct {
}
