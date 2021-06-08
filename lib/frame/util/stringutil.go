package util

import "regexp"

var removeEmptyRowReg *regexp.Regexp = regexp.MustCompile(`(?m)^\s*$\n`)

type stringUtil struct {
}

func (u *stringUtil) RemoveEmptyRow(content string) string {
	return removeEmptyRowReg.ReplaceAllString(content, "")
}

var StringUtil stringUtil = stringUtil{}
