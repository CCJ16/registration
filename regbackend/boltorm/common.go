package boltorm

import (
	"bytes"
	"encoding/gob"
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
