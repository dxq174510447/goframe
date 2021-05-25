package application

import "github.com/dxq174510447/goframe/lib/frame/context"

type ApplicationContextListener interface {
	Starting(local *context.LocalStack)

	EnvironmentPrepared(local *context.LocalStack, environment *ConfigurableEnvironment)

	ContextPrepared(local *context.LocalStack, application *FrameApplicationContext)

	ContextLoaded(local *context.LocalStack, application *FrameApplicationContext)

	Started(local *context.LocalStack, application *FrameApplicationContext)

	Running(local *context.LocalStack, application *FrameApplicationContext)

	Failed(local *context.LocalStack, application *FrameApplicationContext, err error)

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
func (a *ApplicationRunContextListeners) ContextPrepared(local *context.LocalStack, application *FrameApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.ContextPrepared(local, application)
	}
}
func (a *ApplicationRunContextListeners) ContextLoaded(local *context.LocalStack, application *FrameApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.ContextLoaded(local, application)
	}
}
func (a *ApplicationRunContextListeners) Started(local *context.LocalStack, application *FrameApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.Started(local, application)
	}
}
func (a *ApplicationRunContextListeners) Running(local *context.LocalStack, application *FrameApplicationContext) {
	for _, m := range a.ApplicationListeners {
		m.Running(local, application)
	}
}
func (a *ApplicationRunContextListeners) Failed(local *context.LocalStack, application *FrameApplicationContext, err error) {
	for _, m := range a.ApplicationListeners {
		m.Failed(local, application, err)
	}
}

// DefaultApplicationArguments TODO 暂时没使用
type DefaultApplicationArguments struct {
	Args   []string
	Values map[string]string
	isInit bool
}

func (d *DefaultApplicationArguments) parse() {
	if !d.isInit {

		d.isInit = true
	} else {
		return
	}
}

func (d *DefaultApplicationArguments) GetArgByName(key string, defaultValue string) string {
	d.parse()
	if r, ok := d.Values[key]; ok {
		return r
	}
	return defaultValue
}

func (d *DefaultApplicationArguments) GetAllArgs() map[string]string {
	d.parse()
	return d.Values
}

// ConfigurableEnvironment TODO
type ConfigurableEnvironment struct {
	PropertySources *MutablePropertySources
}

func (c *ConfigurableEnvironment) SetActiveProfiles(profiles ...string) {

}

func (c *ConfigurableEnvironment) AddActiveProfiles(profiles string) {

}

func (c *ConfigurableEnvironment) SetDefaultProfiles(profiles ...string) {

}

func (c *ConfigurableEnvironment) GetPropertySources() *MutablePropertySources {
	return c.PropertySources
}

func (c *ConfigurableEnvironment) GetSystemEnvironment() map[string]interface{} {
	return nil
}

func (c *ConfigurableEnvironment) GetSystemProperties() map[string]interface{} {
	return nil
}

func (c *ConfigurableEnvironment) Merge(parent *ConfigurableEnvironment) {

}

type MutablePropertySources struct {
	IsArray  bool
	IsLeaf   bool
	Name     string
	Children *MutablePropertySources
	Raw      string
	Value    string
}
