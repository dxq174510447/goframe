package http

import (
	"github.com/dxq174510447/goframe/lib/frame/context"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"net/http"
	"sort"
)

type FilterChain interface {
	DoFilter(local *context.LocalStack, request *http.Request, response http.ResponseWriter)
}

type Filter interface {
	DoFilter(local *context.LocalStack, request *http.Request, response http.ResponseWriter, chain FilterChain)

	Order() int
}

type DefaultFilterChain struct {
	filters []Filter
}

func (d *DefaultFilterChain) DoFilter(local *context.LocalStack,
	request *http.Request,
	response http.ResponseWriter) {

	index := GetCurrentFilterIndex(local)
	SetCurrentFilterIndex(local, index+1)
	if index >= len(d.filters) {
		//执行 servlet
		GetDispatchServlet().Dispatch(local, request, response)
		return
	}
	d.filters[index].DoFilter(local, request, response, FilterChain(d))

}

var defaultFilterChain DefaultFilterChain = DefaultFilterChain{}

func GetDefaultFilterChain() *DefaultFilterChain {
	return &defaultFilterChain
}

func AddFilter(filter Filter) {

	proxy.AddClassProxy(filter.(proxy.ProxyTarger))

	d := append(defaultFilterChain.filters, filter)
	if len(d) > 1 {
		sort.Slice(d, func(i, j int) bool {
			return d[i].Order() < d[j].Order()
		})
	}
	defaultFilterChain.filters = d
}
