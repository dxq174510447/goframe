package http

import (
	"context"
	"github.com/dxq174510447/goframe/lib/frame/application"
)

type RestControllerAnnotationValue struct {
	Value string
}

type RestControllerAnnotation struct {
	value *RestControllerAnnotationValue
}

func (c *RestControllerAnnotation) AnnotationValue() interface{} {
	return c.value
}

type RestControllerAnnotationSpi struct {
}

func (c *RestControllerAnnotationSpi) AnnotationName(ctx context.Context) string {
	return RestControllerAnnotationName
}

func (c *RestControllerAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &RestControllerAnnotation{
		value: &RestControllerAnnotationValue{},
	}
}

type RequestMappingAnnotationValue struct {
	Value  string
	Method string
}

type RequestMappingAnnotation struct {
	value *RequestMappingAnnotationValue
}

func (c *RequestMappingAnnotation) AnnotationValue() interface{} {
	return c.value
}

type RequestMappingAnnotationSpi struct {
}

func (c *RequestMappingAnnotationSpi) AnnotationName(ctx context.Context) string {
	return RequestMappingAnnotationName
}

func (c *RequestMappingAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &RequestMappingAnnotation{
		value: &RequestMappingAnnotationValue{},
	}
}

type RequestParamAnnotationValue struct {
	Name         string
	DefaultValue string
}

type RequestParamAnnotation struct {
	value *RequestParamAnnotationValue
}

func (r *RequestParamAnnotation) AnnotationValue() interface{} {
	return r.value
}

type RequestParamAnnotationSpi struct {
}

func (r *RequestParamAnnotationSpi) AnnotationName(ctx context.Context) string {
	return RequestParamAnnotationName
}

func (r *RequestParamAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &RequestParamAnnotation{
		value: &RequestParamAnnotationValue{},
	}
}

type RequestHeaderAnnotationValue struct {
	Name         string
	DefaultValue string
}

type RequestHeaderAnnotation struct {
	value *RequestHeaderAnnotationValue
}

func (r *RequestHeaderAnnotation) AnnotationValue() interface{} {
	return r.value
}

type RequestHeaderAnnotationSpi struct {
}

func (r *RequestHeaderAnnotationSpi) AnnotationName(ctx context.Context) string {
	return RequestHeaderAnnotationName
}

func (r *RequestHeaderAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &RequestHeaderAnnotation{
		value: &RequestHeaderAnnotationValue{},
	}
}

type CookieValueAnnotationValue struct {
	Name         string
	DefaultValue string
}

type CookieValueAnnotation struct {
	value *CookieValueAnnotationValue
}

func (r *CookieValueAnnotation) AnnotationValue() interface{} {
	return r.value
}

type CookieValueAnnotationSpi struct {
}

func (r *CookieValueAnnotationSpi) AnnotationName(ctx context.Context) string {
	return CookieValueAnnotationName
}

func (r *CookieValueAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &CookieValueAnnotation{
		value: &CookieValueAnnotationValue{},
	}
}

type RequestBodyAnnotationValue struct {
	//Name string
	//DefaultValue string
}

type RequestBodyAnnotation struct {
	value *RequestBodyAnnotationValue
}

func (r *RequestBodyAnnotation) AnnotationValue() interface{} {
	return r.value
}

type RequestBodyAnnotationSpi struct {
}

func (r *RequestBodyAnnotationSpi) AnnotationName(ctx context.Context) string {
	return RequestBodyAnnotationName
}

func (r *RequestBodyAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &RequestBodyAnnotation{
		value: &RequestBodyAnnotationValue{},
	}
}

type PathVariableAnnotationValue struct {
	Name string
}

type PathVariableAnnotation struct {
	value *PathVariableAnnotationValue
}

func (r *PathVariableAnnotation) AnnotationValue() interface{} {
	return r.value
}

type PathVariableAnnotationSpi struct {
}

func (r *PathVariableAnnotationSpi) AnnotationName(ctx context.Context) string {
	return PathVariableAnnotationName
}

func (r *PathVariableAnnotationSpi) NewAnnotation(ctx context.Context) application.Annotation {
	return &PathVariableAnnotation{
		value: &PathVariableAnnotationValue{},
	}
}

func init() {
	rc := &RestControllerAnnotationSpi{}
	_ = application.AnnotationSpi(rc)
	application.GetResourcePool().RegisterInstance("", rc)

	rm := &RequestMappingAnnotationSpi{}
	_ = application.AnnotationSpi(rm)
	application.GetResourcePool().RegisterInstance("", rm)

	rr := &RequestParamAnnotationSpi{}
	_ = application.AnnotationSpi(rr)
	application.GetResourcePool().RegisterInstance("", rr)

	rb := &RequestBodyAnnotationSpi{}
	_ = application.AnnotationSpi(rb)
	application.GetResourcePool().RegisterInstance("", rb)

	cv := &CookieValueAnnotationSpi{}
	_ = application.AnnotationSpi(cv)
	application.GetResourcePool().RegisterInstance("", cv)

	pv := &PathVariableAnnotationSpi{}
	_ = application.AnnotationSpi(pv)
	application.GetResourcePool().RegisterInstance("", pv)

	rh := &RequestHeaderAnnotationSpi{}
	_ = application.AnnotationSpi(rh)
	application.GetResourcePool().RegisterInstance("", rh)
}
