package event

import (
	"github.com/dxq174510447/goframe/lib/frame/application"
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/log/logclass"
	"github.com/dxq174510447/goframe/lib/frame/proxy/proxyclass"
	"github.com/dxq174510447/goframe/lib/frame/util"
	"reflect"
	"sort"
)

type FrameEventer interface {
	GetSource() interface{}
}

type FrameListener interface {
	OnEvent(local *context.LocalStack, event FrameEventer) error

	Order() int

	WatchEvent() FrameEventer
}

type FrameEventDispatcher struct {
	listeners map[string][]FrameListener

	Logger logclass.AppLoger `FrameAutowired:""`
}

func (f *FrameEventDispatcher) DispatchEvent(local *context.LocalStack, event FrameEventer) {
	eventname := util.ClassUtil.GetClassNameByType(reflect.TypeOf(event).Elem())
	if listeners, ok := f.listeners[eventname]; ok {
		f.Logger.Debug(local, " event %s has %d listener", eventname, len(listeners))
		for _, listener := range listeners {
			listener.OnEvent(local, event)
		}
	} else {
		f.Logger.Debug(local, " event %s has no listener", eventname)
	}
}
func (f *FrameEventDispatcher) AddEventListener(local *context.LocalStack, listener FrameListener) {

	//m,_ := reflect.TypeOf(listener).MethodByName("OnEvent")
	//n := m.Type.NumIn()
	//
	//eventname := util.ClassUtil.GetClassNameByType(m.Type.In(n-1).Elem())
	event := listener.WatchEvent()
	eventname := util.ClassUtil.GetClassNameByType(reflect.TypeOf(event).Elem())
	if f.Logger.IsDebugEnable() {
		listenername := util.ClassUtil.GetClassNameByType(reflect.TypeOf(listener).Elem())
		f.Logger.Debug(local, " listener %s listen %s", listenername, eventname)
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

func (f *FrameEventDispatcher) ProxyTarget() *proxyclass.ProxyClass {
	return nil
}

var frameEventDispatcher FrameEventDispatcher = FrameEventDispatcher{
	listeners: make(map[string][]FrameListener),
}

func GetFrameEventDispatcher() *FrameEventDispatcher {
	return &frameEventDispatcher
}

func init() {
	application.AddProxyInstance("", proxyclass.ProxyTarger(&frameEventDispatcher))
}
