package proxyclass

import (
	"reflect"
	"time"
)

const (
	AnnotationService = "AnnotationService__"
	AnnotationDao     = "AnnotationDao__"
)

var GoTimeType reflect.Type = reflect.TypeOf(time.Time{})
