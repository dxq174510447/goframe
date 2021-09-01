package event

import (
	"context"
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
	"sort"
)

type FrameEventer interface {
	GetSource() interface{}
}

type FrameListener interface {
	OnEvent(local context.Context, event FrameEventer) error

	Order() int

	WatchEvent() FrameEventer
}

type FrameEventDispatcher struct {
	listeners map[string][]FrameListener

	logger application.AppLoger `FrameAutowired:""`
}

func (f *FrameEventDispatcher) DispatchEvent(local context.Context, event FrameEventer) {
	eventname := util.ClassUtil.GetClassNameByType(reflect.TypeOf(event).Elem())
	if listeners, ok := f.listeners[eventname]; ok {
		f.logger.Debug(local, " event %s has %d listener", eventname, len(listeners))
		for _, listener := range listeners {
			listener.OnEvent(local, event)
		}
	} else {
		f.logger.Debug(local, " event %s has no listener", eventname)
	}
}
func (f *FrameEventDispatcher) AddEventListener(local context.Context, listener FrameListener) {

	event := listener.WatchEvent()
	eventname := util.ClassUtil.GetClassNameByType(reflect.TypeOf(event).Elem())

	if f.logger.IsDebugEnable() {
		listenername := util.ClassUtil.GetClassNameByType(reflect.TypeOf(listener).Elem())
		f.logger.Debug(local, " listener %s listen %s", listenername, eventname)
	}

	if listeners, ok := f.listeners[eventname]; ok {
		l := append(listeners, listener)
		sort.Slice(l, func(i, j int) bool {
			return l[i].Order() < l[j].Order()
		})
		f.listeners[eventname] = l
	} else {
		f.listeners[eventname] = []FrameListener{listener}
	}
}

var frameEventDispatcher FrameEventDispatcher = FrameEventDispatcher{
	listeners: make(map[string][]FrameListener),
}

func GetFrameEventDispatcher() *FrameEventDispatcher {
	return &frameEventDispatcher
}

func init() {
	application.GetResourcePool().RegisterInstance("", &frameEventDispatcher)
}
