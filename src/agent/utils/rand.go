package utils

import (
	"time"
)

var (
	x0  uint32 = uint32(time.Now().UnixNano())
	a   uint32 = 1664525
	c   uint32 = 1013904223
	LCG chan uint32
)

const (
	PRERNG = 1024
)

// 全局快速随机数发生器，比标准库快，简单，可预生成
func init() {
	LCG = make(chan uint32, PRERNG)
	go func() {
		for {
			x0 = a*x0 + c
			LCG <- x0
		}
	}()
}
