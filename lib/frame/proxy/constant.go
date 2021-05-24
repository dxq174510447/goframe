package proxy

import "github.com/dxq174510447/goframe/lib/frame/context"

const (
	AnnotationService = "AnnotationService__"
	AnnotationDao     = "AnnotationDao__"
)

const (
	FrameEnvironmentKey = "FrameEnvironmentKey_"
)

func SetEnvironmentToApplication(local *context.LocalStack, env *ConfigurableEnvironment) {
	local.Set(FrameEnvironmentKey, env)
}

// GetEnvironmentFromApplication 如果上下文已经有环境上下文 就使用上一次
func GetEnvironmentFromApplication(local *context.LocalStack) *ConfigurableEnvironment {
	m := local.Get(FrameEnvironmentKey)
	if m == nil {
		return nil
	}
	return m.(*ConfigurableEnvironment)
}
