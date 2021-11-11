package http

import "goframe/lib/frame/application"

type RestControllerAnnotation struct {
	Value string
}

func (c *RestControllerAnnotation) AnnotationValue() interface{} {
	return c.Value
}

type RestControllerAnnotationSpi struct {
}

func (c *RestControllerAnnotationSpi) AnnotationName() string {
	return RestControllerAnnotationName
}

func (c *RestControllerAnnotationSpi) NewAnnotation() application.Annotation {
	return &RestControllerAnnotation{}
}

type RequestMappingAnnotation struct {
	Value  string
	Method string
}

func (c *RequestMappingAnnotation) AnnotationValue() interface{} {
	return c.Value
}

type RequestMappingAnnotationSpi struct {
}

func (c *RequestMappingAnnotationSpi) AnnotationName() string {
	return RequestMappingAnnotationName
}

func (c *RequestMappingAnnotationSpi) NewAnnotation() application.Annotation {
	return &RequestMappingAnnotation{}
}

type RequestParamAnnotation struct {
	Name         string
	DefaultValue string
}

func (r *RequestParamAnnotation) AnnotationValue() interface{} {
	return r.DefaultValue
}

type RequestParamAnnotationSpi struct {
}

func (r *RequestParamAnnotationSpi) AnnotationName() string {
	return RequestParamAnnotationName
}

func (r *RequestParamAnnotationSpi) NewAnnotation() application.Annotation {
	return &RequestParamAnnotation{}
}

type RequestHeaderAnnotation struct {
	Name         string
	DefaultValue string
}

func (r *RequestHeaderAnnotation) AnnotationValue() interface{} {
	return r.DefaultValue
}

type RequestHeaderAnnotationSpi struct {
}

func (r *RequestHeaderAnnotationSpi) AnnotationName() string {
	return RequestHeaderAnnotationName
}

func (r *RequestHeaderAnnotationSpi) NewAnnotation() application.Annotation {
	return &RequestHeaderAnnotation{}
}

type CookieValueAnnotation struct {
	Name         string
	DefaultValue string
}

func (r *CookieValueAnnotation) AnnotationValue() interface{} {
	return r.DefaultValue
}

type CookieValueAnnotationSpi struct {
}

func (r *CookieValueAnnotationSpi) AnnotationName() string {
	return CookieValueAnnotationName
}

func (r *CookieValueAnnotationSpi) NewAnnotation() application.Annotation {
	return &CookieValueAnnotation{}
}

type RequestBodyAnnotation struct {
}

func (r *RequestBodyAnnotation) AnnotationValue() interface{} {
	return nil
}

type RequestBodyAnnotationSpi struct {
}

func (r *RequestBodyAnnotationSpi) AnnotationName() string {
	return RequestBodyAnnotationName
}

func (r *RequestBodyAnnotationSpi) NewAnnotation() application.Annotation {
	return &RequestBodyAnnotation{}
}

type PathVariableAnnotation struct {
	Name string
}

func (r *PathVariableAnnotation) AnnotationValue() interface{} {
	return r.Name
}

type PathVariableAnnotationSpi struct {
}

func (r *PathVariableAnnotationSpi) AnnotationName() string {
	return PathVariableAnnotationName
}

func (r *PathVariableAnnotationSpi) NewAnnotation() application.Annotation {
	return &PathVariableAnnotation{}
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
