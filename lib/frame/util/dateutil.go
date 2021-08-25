package util

import "time"

type dateUtil struct {
}

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
