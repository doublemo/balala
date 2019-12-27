// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package types 类型处理
package types

import (
	"math"
	"reflect"
)

// HasElem 检查切片中是否存在某个值
func HasElem(s, elem interface{}) bool {
	arrV := reflect.ValueOf(s)
	if arrV.Kind() == reflect.Slice || arrV.Kind() == reflect.Array {
		for i := 0; i < arrV.Len(); i++ {
			if reflect.DeepEqual(elem, arrV.Index(i).Interface()) == true {
				return true
			}
		}
	}
	return false
}

// Round 四舍五入保留小数点位数
func Round(f float64, n int) float64 {
	pow10N := math.Pow10(n)
	return math.Trunc((f+0.5/pow10N)*pow10N) / pow10N
}
