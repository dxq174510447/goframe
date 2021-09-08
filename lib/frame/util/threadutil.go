package util

import (
	"context"
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/ctx"
	//"github.com/dxq174510447/goframe/lib/frame/ctx"
	//"github.com/dxq174510447/goframe/lib/frame/ctx"
)

type threadUtil struct {
}

func (t *threadUtil) SetThread(local context.Context, threadId string) string {

	threadName := threadId
	if threadName == "" {
		threadName = DateUtil.FormatNowByType(DatePattern3)
		threadName = fmt.Sprintf("%s-%s", threadName, StringUtil.GetRandomStr(5))
	}
	ctx.WithValue(local, ThreadLocalIdKey, threadName)
	return threadName

}

func (f *threadUtil) GetThread(local context.Context) string {
	m := local.Value(ThreadLocalIdKey)
	if m == nil {
		return ""
	}
	return m.(string)
}

var ThreadUtil threadUtil = threadUtil{}
