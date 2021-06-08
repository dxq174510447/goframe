package util

import (
	"encoding/json"
	"github.com/dxq174510447/goframe/lib/frame/vo"
)

type jsonUtil struct {
}

const WebSuccess int = 0
const WebFailure int = 500

// BuildJsonSuccess 构建成功返回
func (c *jsonUtil) BuildJsonSuccess(r interface{}) *vo.JsonResult {
	var result vo.JsonResult = vo.JsonResult{}

	result.Code = WebSuccess
	result.Data = r

	return &result
}

// BuildJsonFailure 构建失败返回
func (c *jsonUtil) BuildJsonFailure(code int, message string, r interface{}) *vo.JsonResult {
	var result vo.JsonResult = vo.JsonResult{}

	result.Code = code
	result.Data = r
	result.Message = message

	return &result
}

// BuildJsonFailure1 构建失败返回
func (c *jsonUtil) BuildJsonFailure1(message string, r interface{}) *vo.JsonResult {
	var result vo.JsonResult = vo.JsonResult{}

	result.Code = WebFailure
	result.Data = r
	result.Message = message

	return &result
}

// BuildJsonArraySuccess 构建返回数组 例如分页查询
func (c *jsonUtil) BuildJsonArraySuccess(r interface{}, total int) *vo.JsonResult {
	var result vo.JsonResult = vo.JsonResult{}

	result.Code = WebSuccess

	info := vo.JsonArrayResult{Count: total, Data: r}

	result.Data = info

	return &result
}

func (c *jsonUtil) To2String(r interface{}) string {
	result, er := json.Marshal(r)
	if er != nil {
		panic(er)
	}
	return string(result)
}

var JsonUtil jsonUtil = jsonUtil{}
