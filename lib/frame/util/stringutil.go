package util

import (
	"math/rand"
	"regexp"
)

var removeEmptyRowReg *regexp.Regexp = regexp.MustCompile(`(?m)^\s*$\n`)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type stringUtil struct {
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
