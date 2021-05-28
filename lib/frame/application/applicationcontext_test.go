package application

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"testing"
)

type RunFirst struct {
}

var yaml string = `
server:
  port: ${APPLICATION_PORT:8081}
  servlet:
    contextPath: ${APPLICATION_PATH:/api/v1/base}
  access:
    inner: ${APPLICATION_INNER:http://127.0.0.1:8081}
    outer: ${APPLICATION_OUTER:https://wx.dev.chelizitech.com}
spring:
  application:
    name: ${APPLICATION_NAME:base-frontend}
  profiles:
    include: platform
platform:
  baseServer: ${APPLICATION_PATH:/api/v1/base}
  datasource:
    config:
      default:
        username: ${DB_USER:root}
        password: ${DB_PWD:gsta2012!@#}
        url: jdbc:mysql://${DB_HOST:rm-bp1thh63s5tx33q0kio.mysql.rds.aliyuncs.com}:${DB_PORT:3306}/${DB_NAME:plat_base1}?characterEncoding=UTF-8
        driverClass: com.mysql.jdbc.Driver
uploadPath: ${UPLOAD_PATH:d:/img}
`

var yamlDev string = `
server:
  access:
    inner: "aaaaaa"
    outer: "bbbbbb"
`

var platformYaml string = `
spring:
  tempDir: ${APPLICATION_TEMP_DIR:/data/tmp}
  sleuth:
    sampler:
      probability: 100
  zipkin:
    sender:
      type: kafka
  kafka:
    bootstrapServers: ${APPLICATION_KAFKA_URL:47.105.61.26:9092} 
  redis:
##    cluster:
##      nodes: ${APPLICATION_REDIS_CLUSTER_URL:193.112.250.112:7000,193.112.250.112:7001,193.112.250.112:7002}
    host: ${APPLICATION_REDIS_IP:47.105.61.26}
    port: ${APPLICATION_REDIS_PORT:6379}
    password: ${APPLICATION_REDIS_PASSWORD:gsta2012}
    timeout: ${APPLICATION_REDIS_TIMEOUT:5000}
  data:
    mongodb:
      host: ${APPLICATION_MONGODB_IP:47.105.61.26}
      port: ${APPLICATION_MONGODB_PORT:27017}
      database: ${APPLICATION_MONGODB_DATABASE:application}
  cloud:
    httptoken:
      headerName: ${APPLICATION_CLOUD_TOKEN_KEY:token}
      tokenPlatform: ${APPLICATION_CLOUD_TOKEN_PLATFORM:base-frontend}
    circuit:
      breaker:
        enabled: true
    zookeeper:
      connectString: ${APPLICATION_ZOOK_URL:47.105.61.26:2181}
      discovery:
        preferIpAddress: true
        root: /platform/services
      lock: /platform/lock
      coordination: /platform/coordination
ali:
  log:
    accessKey: ${APPLICATION_ALI_LOG_KEY:LTAI4FhakQ36Yg6au6ZDPgsw}
    accessKeySecret: ${APPLICATION_ALI_LOG_SECRET:Ieo6QuXNKRYXdQ9hgvHpwFvMxp5db3}
    logStore: 
      globalSeach: 
        endpoint: null
    properties:
      ioThreadCount: 3
      totalSizeInBytes: 1048576
feign:
  client:
    config:
      default:
        connectTimeout: 10000
        readTimeout: 20000
        loggerLevel: FULL
        retryer: cloud.ecosphere.yy.framework.platform.feignclient.strategy.NeverRetryer
`

type TestAppListener struct {
}

func (t *TestAppListener) Starting(local *context.LocalStack) {
	fmt.Println("TestAppListener", "Starting")
}

func (t *TestAppListener) EnvironmentPrepared(local *context.LocalStack, environment *ConfigurableEnvironment) {
	fmt.Println("TestAppListener", "EnvironmentPrepared")
}

func (t *TestAppListener) Running(local *context.LocalStack, application *FrameApplicationContext) {
	fmt.Println("TestAppListener", "Running")
}

func (t *TestAppListener) Failed(local *context.LocalStack, application *FrameApplicationContext, err interface{}) {
	fmt.Println("TestAppListener", "Failed")
}

func (t *TestAppListener) Order() int {
	return 1
}

func (t *TestAppListener) ProxyTarget() *proxy.ProxyClass {
	return nil
}

func TestApplicationRun(t *testing.T) {
	AddConfigYaml(ApplicationDefaultYaml, yaml)
	AddConfigYaml(ApplicationLocalYaml, yamlDev)
	AddConfigYaml("platform", platformYaml)
	AddProxyInstance("", &TestAppListener{})

	args := []string{}
	app := NewApplication(&RunFirst{})
	app.Run(args)
	fmt.Println(app.Environment.GetBaseValue("ali.log.accessKey", ""))
	//fmt.Println(get)
}
