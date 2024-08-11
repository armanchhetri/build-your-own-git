package packextractor

import (
	"bytes"
	"fmt"
)

type OffDeltaObject struct {
	size      uint64
	deltaData []byte
	offset    uint64
}

type RefDeltaObject struct {
	size    uint64
	refName [20]byte
	refData []byte
}

func (ofd *OffDeltaObject) Raw() ([]byte, error) {
	return ofd.deltaData, nil
}

func (ord *RefDeltaObject) Raw() ([]byte, error) {
	return ord.refData, nil
}

func EmitAOffDelta(stream *bytes.Reader) (Object, error) {
	nextByte, err := stream.ReadByte()
	if err != nil {
		return nil, err
	}
	offset, err := extractVarInt(nextByte, stream)
	if err != nil {
		return nil, err
	}
	data, err := zlibDeflate(stream)
	if err != nil {
		return nil, err
	}
	fmt.Println("OffsetDelta: ", offset, string(data))
	return &OffDeltaObject{deltaData: data, offset: offset}, nil
}

func EmitARefDelta(stream *bytes.Reader) (Object, error) {
	var refName [20]byte
	_, err := stream.Read(refName[:])
	if err != nil {
		return nil, fmt.Errorf("Error while reading refdelta refName %v", err)
	}

	data, err := zlibDeflate(stream)
	if err != nil {
		return nil, err
	}
	fmt.Println("RefsDelta: \n", string(data))
	return &RefDeltaObject{refData: data, refName: refName}, nil
}
