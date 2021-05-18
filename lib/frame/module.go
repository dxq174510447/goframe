package frame

import (
	_ "github.com/dxq174510447/goframe/lib/frame/db"
	"github.com/dxq174510447/goframe/lib/frame/http"
	_ "github.com/dxq174510447/goframe/lib/frame/http"
	_ "github.com/dxq174510447/goframe/lib/frame/proxy"
)

func SetDefaultServletPath(servletPath string) string {
	http.DefaultServletPath = servletPath
	return servletPath
}

func init() {

}
