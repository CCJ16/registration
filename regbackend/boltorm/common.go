package boltorm

import (
	"bytes"
	"encoding/gob"
	"reflect"
)

func encodeData(data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

func decodeData(buf []byte, data interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(buf))
	return decoder.Decode(data)
}

func makeSliceFor(dataType interface{}) interface{} {
	return reflect.New(reflect.SliceOf(reflect.TypeOf(dataType))).Elem().Interface()
}

func makeNew(dataType interface{}) interface{} {
	return reflect.New(reflect.TypeOf(dataType)).Interface()
}

func appendToSlice(slice interface{}, nextElement interface{}) interface{} {
	return reflect.Append(reflect.ValueOf(slice), reflect.ValueOf(nextElement).Elem()).Interface()
}
