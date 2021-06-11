package util

import "time"

type dateUtil struct {
}

var DatePattern1 string = "2006-01-02 15:04:05"
var DatePattern2 string = "20060102150405"

func (d *dateUtil) FormatNow() string {
	return time.Now().Format(DatePattern1)
}
func (d *dateUtil) FormatNowByType(pattern string) string {
	return time.Now().Format(pattern)
}

var DateUtil dateUtil = dateUtil{}