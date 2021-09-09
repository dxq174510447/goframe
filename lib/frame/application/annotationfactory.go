package application

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"sync"
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

	var annotationVal *yaml.Node
	var annotationName string
	for i := 0; i < len(value.Content); i++ {
		if strings.ToLower(value.Content[i].Value) == "name" {
			annotationName = value.Content[i+1].Value
		}
		if strings.ToLower(value.Content[i].Value) == "value" {
			annotationVal = value.Content[i+1]
		}
	}

	if annotationName == "" {
		panic(fmt.Errorf("not find annotation name"))
	}

	anno := GetAnnotationFactory().NewAnnotation(annotationName)

	if anno == nil {
		panic(fmt.Errorf("%s not find annotation type", annotationName))
	}

	t.Name = annotationName
	t.Value = anno

	err := annotationVal.Decode(anno)

	if err != nil {
		return err
	}

	return nil
}

type AnnotationFactory struct {
	logger   AppLoger
	spiMap   map[string]AnnotationSpi
	initLock sync.Once
}

func (a *AnnotationFactory) init() {
	a.initLock.Do(func() {
		if a.spiMap == nil {
			a.spiMap = make(map[string]AnnotationSpi)
		}
	})
}

func (a *AnnotationFactory) NewAnnotation(annotationName string) Annotation {
	a.init()
	k := strings.ToLower(annotationName)

	if v, ok := a.spiMap[k]; ok {
		return v.NewAnnotation()
	}
	return nil
}

func (a *AnnotationFactory) AddAnnotationSpi(spi AnnotationSpi) {
	a.init()
	a.spiMap[strings.ToLower(spi.AnnotationName())] = spi
}

var annotationFactory *AnnotationFactory
var annotationFactoryLock sync.Once = sync.Once{}

func GetAnnotationFactory() *AnnotationFactory {
	annotationFactoryLock.Do(func() {
		annotationFactory = &AnnotationFactory{}
		//annotationFactory.logger = GetResourcePool().ProxyInsPool.LogFactory.GetLoggerType(reflect.TypeOf(annotationFactory))
	})
	return annotationFactory
}
