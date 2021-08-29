package util

import (
	"fmt"
	"math/rand"
	"strings"
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

func (u *stringUtil) FirstLower(name string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(name[0:1]), name[1:])
}

func (u *stringUtil) FirstUpper(name string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(name[0:1]), name[1:])
}

var StringUtil stringUtil = stringUtil{}
