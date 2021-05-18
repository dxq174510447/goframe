package exception

import "fmt"

// FrameException 可见异常
type FrameException struct {
	Code    int
	Message string
}

func (f *FrameException) Error() string {
	r := fmt.Sprintf("%s , 编号 %d", f.Message, f.Code)
	return r
}

// NewException 创建错误异常
func NewException(code int, message string) *FrameException {
	return &FrameException{Code: code, Message: message}
}
