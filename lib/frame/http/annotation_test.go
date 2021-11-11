package http

import (
	"fmt"
	"goframe/lib/frame/application"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"testing"
)

const outputFile = "/Users/klook/log/test.yaml"
const inputFile = "/Users/klook/log/test.yaml"

func TestGenerateYaml(t *testing.T) {

	//body, err := yaml.Marshal(v1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(body))
	//ioutil.WriteFile(outputFile, body, 0644)0644
}

func TestGenerateObject(t *testing.T) {

	rc := &RestControllerAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(rc)
	rm := &RequestMappingAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(rm)
	rr := &RequestParamAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(rr)
	rb := &RequestBodyAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(rb)
	cv := &CookieValueAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(cv)
	pv := &PathVariableAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(pv)
	rh := &RequestHeaderAnnotationSpi{}
	application.GetAnnotationFactory().AddAnnotationSpi(rh)

	body, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	var v2 []*application.ClassV1
	err = yaml.Unmarshal(body, &v2)
	if err != nil {
		panic(err)
	}
	fmt.Println(v2)
}
