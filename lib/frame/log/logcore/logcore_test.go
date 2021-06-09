package logcore

import "testing"

var m1 string = `
<configuration scan="true" scanPeriod="30 seconds"  packagingData="true"> 
  <appender name="ERROR_FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
  	<filter class="ch.qos.logback.classic.filter.LevelFilter">
      <level>ERROR</level>
      <onMatch>ACCEPT</onMatch>
      <onMismatch>DENY</onMismatch>
    </filter>
    <file>${LOG}/error.log</file>
    <rollingPolicy class="ch.qos.logback.core.rolling.SizeAndTimeBasedRollingPolicy">
      <!-- rollover daily -->
      <fileNamePattern>${LOG}/error-%d{yyyy-MM-dd}.%i.log</fileNamePattern>
       <!-- each file should be at most 100MB, keep 60 days worth of history, but at most 20GB -->
       <maxFileSize>100MB</maxFileSize>    
       <maxHistory>60</maxHistory>
       <totalSizeCap>20GB</totalSizeCap>
    </rollingPolicy>
    <encoder>
      <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>
  
  
  <appender name="INFO_FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
    <file>${LOG}/info.log</file>
    <rollingPolicy class="ch.qos.logback.core.rolling.SizeAndTimeBasedRollingPolicy">
      <!-- rollover daily -->
      <fileNamePattern>${LOG}/info-%d{yyyy-MM-dd}.%i.log</fileNamePattern>
       <!-- each file should be at most 100MB, keep 60 days worth of history, but at most 20GB -->
       <maxFileSize>100MB</maxFileSize>    
       <maxHistory>60</maxHistory>
       <totalSizeCap>20GB</totalSizeCap>
    </rollingPolicy>
    <encoder>
      <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>
  
  <appender name="CONSOLE" class="console">
    <encoder>
      <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>

  <logger name="org" level="DEBUG">
			<appender-ref ref="CONSOLE"/>
			<appender-ref ref="INFO_FILE"/>
	</logger>
	<logger name="cloud" level="DEBUG">
				<appender-ref ref="CONSOLE"/>
				<appender-ref ref="INFO_FILE"/>
	</logger>
  
  <logger name="org.mybatis" level="ERROR"/>
	<logger name="org.mybatis.framework.sql" level="ERROR"/>
  <logger name="org.springframework" level="ERROR"/>
  <logger name="org.quartz" level="ERROR"/>
  <logger name="org.apache" level="ERROR"/>
  <logger name="org.mongodb" level="ERROR"/>
  <root level="ERROR">
  </root>
</configuration>
`

func TestName(t *testing.T) {
	l := &LogFactory{}
	l.Parse(m1, nil)
}
