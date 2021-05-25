package application

import "testing"

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

func TestApplicationRun(t *testing.T) {
	AddConfigYaml("application.yml", yaml)
	AddConfigYaml("application-platform.yml", yamlDev)
}
