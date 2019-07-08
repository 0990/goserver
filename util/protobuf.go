package util

import (
	"github.com/golang/protobuf/proto"
	"reflect"
)

//TODO 比较msgType.String()和msgType.Elem().Name()区别
//func ProtoHash(msg proto.Message) (uint16, reflect.Type) {
//	msgType := reflect.TypeOf(msg)
//	return StringHash(msgType.String()), reflect.TypeOf(msg)
//}

func ProtoHash(msg proto.Message) (uint32, reflect.Type) {
	msgType := reflect.TypeOf(msg)
	return CRC32Hash(msgType.String()), reflect.TypeOf(msg)
}
