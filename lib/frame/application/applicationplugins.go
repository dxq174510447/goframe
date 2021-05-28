package application

import (
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"os"
	"regexp"
	"strings"
)

type ApplicationContextListener interface {
	Starting(local *context.LocalStack)

	EnvironmentPrepared(local *context.LocalStack, environment *ConfigurableEnvironment)

	Running(local *context.LocalStack, application *FrameApplicationContext)

	Failed(local *context.LocalStack, application *FrameApplicationContext, err interface{})

	Order() int
}

type ApplicationRunContextListeners struct {
	ApplicationListeners []ApplicationContextListener
	Args                 *DefaultApplicationArguments
}

func (a *ApplicationRunContextListeners) Starting(local *context.LocalStack) {
	for _, m := range a.ApplicationListeners {
		m.Starting(local)
	}
}
func (a *ApplicationRunContextListeners) EnvironmentPrepared(local *context.LocalStack, environment *ConfigurableEnvironment) {
	for _, m := range a.ApplicationListeners {
		m.EnvironmentPrepared(local, environment)
	}
}
func (a *ApplicationRunContextListeners) Running(local *context.LocalStack, application *FrameApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.Running(local, application)
	}
}
func (a *ApplicationRunContextListeners) Failed(local *context.LocalStack, application *FrameApplicationContext, err interface{}) {
	for _, m := range a.ApplicationListeners {
		m.Failed(local, application, err)
	}
}

type DefaultApplicationArguments struct {
	Args   []string
	ArgMap map[string]string
}

func (d *DefaultApplicationArguments) Parse() {

	if d.ArgMap == nil {
		d.ArgMap = make(map[string]string)
	}

	if len(d.Args) == 0 {
		return
	}

	reg := regexp.MustCompile(`^\\-+`)
	for _, arg := range d.Args {
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
		d.ArgMap[strings.TrimSpace(k1)] = strings.TrimSpace(v1)
	}
}

func (d *DefaultApplicationArguments) GetByName(key string, defaultValue string) string {
	if m, ok := d.ArgMap[key]; ok {
		return m
	}
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	v := os.Getenv(envKey)
	if v != "" {
		return v
	}
	return defaultValue
}

type ConfigurableEnvironment struct {
	ConfigTree *YamlTree
	AppArgs    *DefaultApplicationArguments
}

func (y *ConfigurableEnvironment) Parse(content string) {
	if y.ConfigTree.AppArgs == nil {
		y.ConfigTree.AppArgs = y.AppArgs
	}
	y.ConfigTree.Parse(content)
}

func (y *ConfigurableEnvironment) GetBaseValue(key string, defaultValue string) string {
	m := y.ConfigTree.GetBaseValue(key)
	if m == "" {
		return defaultValue
	}
	return m
}

// FrameLoadInstanceHandler 自定义实例加载模式
type FrameLoadInstanceHandler interface {

	// LoadInstance 返回bool 自定义加载返回true 交给框架默认处理返回false
	LoadInstance(local *context.LocalStack, target proxy.ProxyTarger,
		application *FrameApplication,
		applicationContext *FrameApplicationContext) bool

	Order() int
}
