package dbcore

import "github.com/dxq174510447/goframe/lib/frame/exception"

type DaoException struct {
	exception.FrameException
}

// NewException 创建错误异常
func NewDaoException(code int, message string) *DaoException {
	return &DaoException{exception.FrameException{Code: code, Message: message}}
}
