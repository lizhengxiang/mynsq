package util

import (
	"sync"
)

//
//本文件是对WaitGroup的封装，关于WaitGroup，根据字义，wait是等待的意思，group是组、团体的意思，合起来就是指等待一个组。
//即指，当一个组里所有的操作都完成后，才会继续执行。
//可以参考http://www.baiyuxiong.com/?p=913理解WaitGroup用法
//

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
