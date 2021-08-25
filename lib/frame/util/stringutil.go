package util

import (
	"math/rand"
)

type stringUtil struct {
}

func (u *stringUtil) IsFirstCharUpperCase(content string) bool {
	firstChar := content[0]
	if firstChar < 65 || firstChar > 90 {
		return false
	} else {
		return true
	}
}

func (u *stringUtil) RemoveEmptyRow(content string) string {
	return removeEmptyRowReg.ReplaceAllString(content, "")
}

func (u *stringUtil) GetRandomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

var StringUtil stringUtil = stringUtil{}
