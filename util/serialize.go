package util

import (
	"bytes"
	"encoding/gob"
)

// 序列化对象
func Serialize(data interface{}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(data)
	return buffer.Bytes()
}
