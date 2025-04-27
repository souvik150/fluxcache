package protohelper

import (
	"bytes"
	"compress/gzip"
	"io"


	"google.golang.org/protobuf/proto"
)

func Marshal(msg proto.Message) ([]byte, error) {
	raw, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return compress(raw)
}

func Unmarshal(data []byte, msg proto.Message) error {
	raw, err := decompress(data)
	if err != nil {
		return err
	}
	return proto.Unmarshal(raw, msg)
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}
