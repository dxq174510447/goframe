package application

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"testing"
)

const outputFile = "/Users/klook/log/test.yaml"
const inputFile = "/Users/klook/log/test.yaml"

func TestGenerateYaml(t *testing.T) {

	body, err := yaml.Marshal(v1)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	ioutil.WriteFile(outputFile, body, 0644)
}

func TestGenerateObject(t *testing.T) {

	rc := &RestControllerAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(rc)
	rm := &RequestMappingAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(rm)
	rr := &RequestParamAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(rr)
	rb := &RequestBodyAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(rb)
	cv := &CookieValueAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(cv)
	pv := &PathVariableAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(pv)
	rh := &RequestHeaderAnnotationSpi{}
	GetAnnotationFactory().AddAnnotationSpi(rh)

	body, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	var v2 []*ClassV1
	err = yaml.Unmarshal(body, &v2)
	if err != nil {
		panic(err)
	}
	fmt.Println(v2)
}

var v1 []*ClassV1 = []*ClassV1{
	{
		Name:    "UserController",
		PkgName: "com.kl.controller.user.UserController",
		Annotation: []*TypeAnnotationV1{
			{
				Name: "RestController",
			},
			{
				Name: "RequestMapping",
				Value: &RequestMappingAnnotation{
					Value: "/v1/user",
				},
			},
		},
		Method: []*ClassMethodV1{
			{
				Name: "Save",
				Annotation: []*TypeAnnotationV1{
					{
						Name: "RequestMapping",
						Value: &RequestMappingAnnotation{
							Value:  "/",
							Method: "post",
						},
					},
				},
				In: []*ClassParameterV1{
					{
						Name:    "req",
						PkgName: "com.lang.tring.User",
						Annotation: []*TypeAnnotationV1{
							{
								Name:  "RequestBody",
								Value: &RequestBodyAnnotation{},
							},
						},
					},
				},
				Out: []*ClassParameterV1{
					{
						PkgName: "com.lang.tring.User",
					},
				},
			}, {
				Name: "Update",
				Annotation: []*TypeAnnotationV1{
					{
						Name: "RequestMapping",
						Value: &RequestMappingAnnotation{
							Value:  "/",
							Method: "put",
						},
					},
				},
				In: []*ClassParameterV1{
					{
						Name:    "req",
						PkgName: "com.lang.tring.User",
						Annotation: []*TypeAnnotationV1{
							{
								Name:  "RequestBody",
								Value: &RequestBodyAnnotation{},
							},
						},
					}, {
						Name:    "id",
						PkgName: "com.lang.tring.string",
						Annotation: []*TypeAnnotationV1{
							{
								Name: "RequestParam",
								Value: &RequestParamAnnotation{
									Name: "id",
								},
							},
						},
					},
				},
				Out: []*ClassParameterV1{
					{
						PkgName: "com.lang.tring.User",
					},
				},
			},
		},
	}}

var annotation_config string = `
lib:
  frame:
    application:
      UserController:
        annotation:
        - name: "UserRestController"
          value: ""
        - name: "RequestMapping"
          value:
            value: "/user"
        method:
          Save:
            annotation:
            - name: "RequestMapping"
              value:
                value: "/user"
                method: "get"
            in:
            - name: "body"
              type: "go.lang.string" 
              annotation:
              - name: "RequestBody"
            out:
            - name: "body"
              type: "go.lang.string" 
              annotation:
              - name: "RequestBody"
            
`
