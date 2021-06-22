package util

import (
	"math"
	"strconv"
)

type mathUtil struct {
}

func (m *mathUtil) Str2int(str string) (int, error) {
	if str == "" {
		return 0, nil
	}
	return strconv.Atoi(str)
}

func (m *mathUtil) Str2int64(str string) (int64, error) {
	if str == "" {
		return int64(0), nil
	}
	return strconv.ParseInt(str, 10, 64)
}

func (m *mathUtil) Str2float64(str string) (float64, error) {
	if str == "" {
		return float64(0.0), nil
	}
	return strconv.ParseFloat(str, 64)
}

func (m *mathUtil) Round(n1 float64, scale int) (float64, error) {
	if scale == 0 {
		return math.Round(n1), nil
	}
	t := math.Pow10(scale)
	r1 := math.Round(n1*t) / t
	return r1, nil
}

func (m *mathUtil) Floor(n1 float64, scale int) (float64, error) {
	if scale == 0 {
		return math.Floor(n1), nil
	}
	t := math.Pow10(scale)
	r1 := math.Floor(n1*t) / t
	return r1, nil
}

func (m *mathUtil) Ceil(n1 float64, scale int) (float64, error) {
	if scale == 0 {
		return math.Ceil(n1), nil
	}
	t := math.Pow10(scale)
	r1 := math.Ceil(n1*t) / t
	return r1, nil
}

var MathUtil mathUtil = mathUtil{}
