package context

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/util"
)

// LocalStack 针对上下文设置的环境变量 嵌套 -----------------
type LocalStack struct {
	element []map[string]interface{}
}

// Push 新增层次
func (f *LocalStack) Push() map[string]interface{} {
	ele := make(map[string]interface{})
	if f.element == nil {
		f.element = []map[string]interface{}{ele}
	} else {
		f.element = append(f.element, ele)
	}
	return ele
}

// Pop 出栈
func (f *LocalStack) Pop() map[string]interface{} {
	result := f.element[len(f.element)-1]
	f.element = f.element[0 : len(f.element)-1]
	return result
}

// Peek 查看栈顶的环境设置
func (f *LocalStack) Peek() map[string]interface{} {
	return f.element[len(f.element)-1]
}

// Set 在栈顶环境变量设置参数
func (f *LocalStack) Set(key string, value interface{}) {
	top := f.Peek()
	top[key] = value
}

func (f *LocalStack) Destroy() {
	f.element = nil
}

// Get 从栈中依次取出环境变量key，从栈顶开始
func (f *LocalStack) Get(key string) interface{} {
	if len(f.element) == 0 {
		return nil
	}
	n := len(f.element) - 1
	for i := n; i >= 0; i-- {
		ele := f.element[i]
		if res, ok := ele[key]; ok {
			return res
		}
	}
	return nil
}

func (f *LocalStack) SetThread() string {
	threadName := util.DateUtil.FormatNowByType(util.DatePattern3)
	threadName = fmt.Sprintf("%s-%s", threadName, util.StringUtil.GetRandomStr(5))
	top := f.Peek()
	top[ThreadLocalIdKey] = threadName
	return threadName
}

func (f *LocalStack) GetThread() string {
	m := f.Get(ThreadLocalIdKey)
	if m == nil {
		return ""
	}
	return m.(string)
}

// NewLocalStack 创建新的变量栈
func NewLocalStack() *LocalStack {
	result := &LocalStack{}
	result.Push()
	return result
}
