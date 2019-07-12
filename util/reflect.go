package util

import (
	"errors"
	"reflect"
)

//检查形如 func(args0 xx,args1 proto.Message)这种的func
func CheckArgs1MsgFun(cb interface{}) (err error, funValue reflect.Value, msgType reflect.Type) {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		err = errors.New("cb not a func")
		return
	}

	numArgs := cbType.NumIn()
	if numArgs != 2 {
		err = errors.New("cb param num args !=2")
		return
	}

	//TODO 严格检查参数类型
	//args0 := cbType.In(0)
	msgType = cbType.In(1)
	if msgType.Kind() != reflect.Ptr {
		err = errors.New("cb param args1 not ptr")
		return
	}

	funValue = reflect.ValueOf(cb)
	return
}
