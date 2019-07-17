package util

import "runtime/debug"

func ProtectedFun(f func()) {
	//TODO 每个函数都包装了defer，性能怎样？
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	f()
}
