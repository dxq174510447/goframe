package application

import (
	"context"
	"reflect"
)

type AnnotationFactory struct {
	logger AppLoger
}

func (a *AnnotationFactory) NewAnnotation(ctx context.Context) Annotation {
	return nil
}

func NewAnnotationFactory() *AnnotationFactory {
	f := &AnnotationFactory{}
	f.logger = GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(reflect.TypeOf(f))
	return f
}
