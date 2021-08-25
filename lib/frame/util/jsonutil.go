package util

import (
	"encoding/json"
)

type jsonUtil struct {
}

func (c *jsonUtil) To2String(r interface{}) string {
	result, er := json.Marshal(r)
	if er != nil {
		panic(er)
	}
	return string(result)
}

var JsonUtil jsonUtil = jsonUtil{}
