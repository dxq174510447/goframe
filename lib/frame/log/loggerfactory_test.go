package log

import (
	"testing"
)

var defaultLogbackTest = `
<configuration> 
  
  <appender name="console1" class="console">
    <encoder>
      <pattern>%date [%-5thread] [%-5level] %-10logger{0}.%-5line %msg %n</pattern>
    </encoder>
  </appender>

<appender name="console2" class="console">
    <encoder>
      <pattern>%date [%-5thread] [%-5level] %-10logger{0}.%-5line %msg %n</pattern>
    </encoder>
  </appender>
  
  <logger name="test/a/b" level="debug">
	<appender-ref ref="console1"/>
  </logger>

  <logger name="test/a/b/c/d" level="debug">
	<appender-ref ref="console1"/>
  </logger>

  <logger name="test/b/b/c/d" level="debug">
	<appender-ref ref="console1"/>
    <appender-ref ref="console2"/>
  </logger>
  <root level="info">
	<appender-ref ref="DEFAULT_CONSOLE"/>
  </root>
</configuration>
`

func TestLoggerFactory_ParseAndReload(t *testing.T) {

	factory := GetLoggerFactory()
	factory.PrintTree()
	factory.ParseAndReload(defaultLogbackTest, nil)
	factory.PrintTree()
	//time.Sleep(time.Second*100)

	//l := sync.Once{}
	//l.Do(func() {
	//	fmt.Println("aa")
	//	fmt.Println("bbbb")
	//})
	//
	//a.init()

}
