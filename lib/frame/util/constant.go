package util

import "regexp"

const ThreadLocalIdKey = "ThreadLocalIdKey_"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const DatePattern1 = "2006-01-02 15:04:05"
const DatePattern2 = "20060102150405"
const DatePattern3 = "0102150405"
const DatePattern4 = "2006-01-02"

var removeEmptyRowReg *regexp.Regexp = regexp.MustCompile(`(?m)^\s*$\n`)
