package structure

//Stack 栈->元素map
type Stack struct {
	element []map[string]interface{}
}

// Push 新增层次
func (f *Stack) Push() map[string]interface{} {
	ele := make(map[string]interface{})
	if f.element == nil {
		f.element = []map[string]interface{}{ele}
	} else {
		f.element = append(f.element, ele)
	}
	return ele
}

// Pop 出栈
func (f *Stack) Pop() map[string]interface{} {
	result := f.element[len(f.element)-1]
	f.element = f.element[0 : len(f.element)-1]
	return result
}

// Peek 查看栈顶的环境设置
func (f *Stack) Peek() map[string]interface{} {
	return f.element[len(f.element)-1]
}

// Set 在栈顶环境变量设置参数
func (f *Stack) Set(key string, value interface{}) {
	top := f.Peek()
	top[key] = value
}

func (f *Stack) Destroy() {
	f.element = nil
}

// Get 从栈中依次取出环境变量key，从栈顶开始
func (f *Stack) Get(key string) interface{} {
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

// NewLocalStack 创建新的变量栈
func NewStack() *Stack {
	result := &Stack{}
	result.Push()
	return result
}
