package util

import "time"

type dateUtil struct {
}

var DatePattern1 string = "2006-01-02 15:04:05"
var DatePattern2 string = "20060102150405"
var DatePattern3 string = "0102150405"
var DatePattern4 string = "2006-01-02"

func (d *dateUtil) FormatNow() string {
	return time.Now().Format(DatePattern1)
}
func (d *dateUtil) FormatNowByType(pattern string) string {
	return time.Now().Format(pattern)
}
func (d *dateUtil) FormatByType(time2 *time.Time, pattern string) string {
	return time2.Format(pattern)
}

func (d *dateUtil) Cover2Time(time1 string, pattern string) (*time.Time, error) {
	t, err := time.Parse(pattern, time1)
	return &t, err
}

var DateUtil dateUtil = dateUtil{}
