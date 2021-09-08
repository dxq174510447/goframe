package application

import (
	"context"
	"gopkg.in/yaml.v3"
	"reflect"
)

type ClassV1 struct {
	Name       string
	PkgName    string
	Annotation []*TypeAnnotationV1
	Method     []*ClassMethodV1
}

type ClassMethodV1 struct {
	Name       string
	Annotation []*TypeAnnotationV1
	In         []*ClassParameterV1
	Out        []*ClassParameterV1
}

type ClassParameterV1 struct {
	Name       string
	PkgName    string
	Annotation []*TypeAnnotationV1
}

type TypeAnnotationV1 struct {
	Name  string
	Value Annotation
}

func (t *TypeAnnotationV1) UnmarshalYAML(value *yaml.Node) error {
	return nil
}

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
