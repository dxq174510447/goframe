package application

const (
	AnnotationRestController = "AnnotationRestController_"

	AnnotationController = "AnnotationController_"

	AnnotationValueRestKey = "AnnotationValueRestKey_"

	FilterIndexWaitToExecute = "FilterIndexWaitToExecute_"

	CurrentControllerInvoker = "CurrentControllerInvoker_"

	CurrentHttpRequest = "CurrentHttpRequest_"

	CurrentHttpResponse = "CurrentHttpResponse_"

	RestControllerAnnotationName = "RestController"

	RequestMappingAnnotationName = "RequestMapping"

	RequestParamAnnotationName = "RequestParam"

	RequestBodyAnnotationName = "RequestBody"

	CookieValueAnnotationName = "CookieValue"

	PathVariableAnnotationName = "PathVariable"

	RequestHeaderAnnotationName = "RequestHeader"
)

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

func (c *RestControllerAnnotationSpi) NewAnnotation() Annotation {
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

func (c *RequestMappingAnnotationSpi) NewAnnotation() Annotation {
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

func (r *RequestParamAnnotationSpi) NewAnnotation() Annotation {
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

func (r *RequestHeaderAnnotationSpi) NewAnnotation() Annotation {
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

func (r *CookieValueAnnotationSpi) NewAnnotation() Annotation {
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

func (r *RequestBodyAnnotationSpi) NewAnnotation() Annotation {
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

func (r *PathVariableAnnotationSpi) NewAnnotation() Annotation {
	return &PathVariableAnnotation{}
}

//func init() {
//	rc := &RestControllerAnnotationSpi{}
//	_ = AnnotationSpi(rc)
//	GetResourcePool().RegisterInstance("", rc)
//
//	rm := &RequestMappingAnnotationSpi{}
//	_ = AnnotationSpi(rm)
//	GetResourcePool().RegisterInstance("", rm)
//
//	rr := &RequestParamAnnotationSpi{}
//	_ = AnnotationSpi(rr)
//	GetResourcePool().RegisterInstance("", rr)
//
//	rb := &RequestBodyAnnotationSpi{}
//	_ = AnnotationSpi(rb)
//	GetResourcePool().RegisterInstance("", rb)
//
//	cv := &CookieValueAnnotationSpi{}
//	_ = AnnotationSpi(cv)
//	GetResourcePool().RegisterInstance("", cv)
//
//	pv := &PathVariableAnnotationSpi{}
//	_ = AnnotationSpi(pv)
//	GetResourcePool().RegisterInstance("", pv)
//
//	rh := &RequestHeaderAnnotationSpi{}
//	_ = AnnotationSpi(rh)
//	GetResourcePool().RegisterInstance("", rh)
//}
