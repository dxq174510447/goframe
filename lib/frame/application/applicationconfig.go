package application

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"
)

/**
应用参数和环境变量
*/
type ApplicationArguments struct {
	argMap   map[string]string
	initLock sync.Once
	logger   AppLoger
}

func (d *ApplicationArguments) init() {
	d.initLock.Do(func() {
		d.argMap = make(map[string]string)
	})
}

func (d *ApplicationArguments) Parse(args []string) {

	d.init()

	if len(args) == 0 {
		return
	}

	reg := regexp.MustCompile(`^\\-+`)
	for _, arg := range args {
		arg1 := reg.ReplaceAllString(strings.TrimSpace(arg), "")
		p := strings.Index(arg1, "=")
		var k1 string
		var v1 string
		if p < 0 {
			k1 = arg1
			v1 = ""
		} else {
			k1 = arg1[0:p]
			v1 = arg1[p+1 : len(arg1)]
		}
		d.argMap[strings.TrimSpace(k1)] = strings.TrimSpace(v1)
	}
}

func (d *ApplicationArguments) GetByName(key string, defaultValue string) string {
	if m, ok := d.argMap[key]; ok {
		return m
	}
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	v := os.Getenv(envKey)
	if v != "" {
		return v
	}
	return defaultValue
}

/**
应用配置
*/
type ApplicationConfig struct {
	// 应用配置
	configTree *YamlTree
	// 环境变量和启动参数变量
	appArgs  *ApplicationArguments
	initLock sync.Once
	logger   AppLoger
}

func (a *ApplicationConfig) init() {
	a.initLock.Do(func() {
		if a.configTree == nil {
			a.configTree = NewYamlTree(a.appArgs)
		}
	})
}

//RefreshConfigTree 当merge tree之后刷新一下
func (a *ApplicationConfig) RefreshConfigTree() {
	a.configTree.ReIndex()
}

func (a *ApplicationConfig) SetAppArguments(appArgs *ApplicationArguments) *ApplicationConfig {
	a.appArgs = appArgs
	return a
}

func (a *ApplicationConfig) GetTplFuncMap() template.FuncMap {
	return template.FuncMap{
		"env": func(key string, defaultValue string) string {
			return a.GetBaseValue(key, defaultValue)
		},
	}
}

func (a *ApplicationConfig) Parse(content string) {
	a.init()
	a.configTree.Parse(content)
}

func (y *ApplicationConfig) GetObjectValue(key string, target interface{}) {
	y.configTree.GetObjectValue(key, target)
}

func (y *ApplicationConfig) GetBaseValue(key string, defaultValue string) string {
	m := y.configTree.GetBaseValue(key)
	if m == "" {
		return defaultValue
	}
	return m
}

func NewApplicationConfig(appArgs *ApplicationArguments) *ApplicationConfig {
	ac := &ApplicationConfig{
		appArgs:    appArgs,
		configTree: NewYamlTree(appArgs),
	}
	ac.logger = GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(reflect.TypeOf(ac))
	return ac
}

func NewApplicationArguments() *ApplicationArguments {
	aa := &ApplicationArguments{}
	aa.logger = GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(reflect.TypeOf(aa))
	return aa
}
