package packextractor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
)

type ObjectType uint8

const (
	OBJ_INVALID   ObjectType = 0
	OBJ_COMMIT    ObjectType = 1
	OBJ_TREE      ObjectType = 2
	OBJ_BLOB      ObjectType = 3
	OBJ_TAG       ObjectType = 4
	OBJ_OFS_DELTA ObjectType = 6
	OBJ_REF_DELTA ObjectType = 7
)

type Object interface {
	Raw() ([]byte, error)
}

type ObjectCommit struct {
	currentHash   string
	parentHash    string
	authorName    string
	authorEmail   string
	commiterName  string
	commiterEmail string
	timestamp     string
	message       string
	size          uint
}

func (oc *ObjectCommit) Raw() ([]byte, error) {
	return []byte{}, nil
}

type BlobObject struct {
	content string
	size    uint
}

type TreeLeaf struct {
	filename string
	sha1hash string
	mode     string
}

type ObjectTree = []TreeLeaf

type TagObject struct {
	size        uint
	tagType     string // "commit"| "blob"
	name        string
	sha1hash    string
	taggerName  string
	taggerEmail string
	timestamp   int64
	timezone    string
	message     string
}

func EmitAGitObject(stream *bytes.Reader) (Object, error) {
	nextByte, err := stream.ReadByte()
	if err != nil {
		return nil, err
	}
	const typeMask uint8 = 112 // 0b01110000
	objectType := ObjectType(nextByte & typeMask >> 4)
	size, err := extractVarInt(nextByte, stream)
	if err != nil {
		return nil, fmt.Errorf("Error while extracting size of an object %v", err)
	}
	fmt.Printf("objType: %v, size: %v\n", objectType, size)

	switch objectType {
	case OBJ_COMMIT, OBJ_TREE, OBJ_BLOB, OBJ_TAG:
		return EmitACommitObject(stream)
	case OBJ_OFS_DELTA:
		return EmitAOffDelta(stream)

	case OBJ_REF_DELTA:
		return EmitARefDelta(stream)

	default:
		slog.Error("Invalid object type detected", objectType)
	}
	return nil, nil
}

func extractVarInt(nextByte byte, stream *bytes.Reader) (uint64, error) {
	var sizeMask uint8 = 15             // 0b00001111
	size := uint64(nextByte & sizeMask) // assuming that an object size can be represented in 64bit uint
	shift := 4
	// first 4 bits form least significant byte for integer: size  eg: 0000XXXX, first 4 are 0 and last 4 are actual bits
	// other 7 bits are prepended to first 4 while first bit is 1
	// for eg: byte array [144 15 120] will form integer bit sequence of 0b11110000 = 240
	var err error
	for int(nextByte) > 128 {

		nextByte, err = stream.ReadByte()
		if err != nil {
			return 0, err
		}
		var msbMask uint8 = 127
		currSize := uint64(nextByte & msbMask)
		size = size | currSize<<shift
		shift += 7
		// break
	}

	return size, nil
}

func zlibDeflate(stream *bytes.Reader) ([]byte, error) {
	r, err := zlib.NewReader(stream)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		slog.Error(err.Error())
		return nil, fmt.Errorf("Error on deflating zlib data %v", err)
	}
	return data, nil
}

func EmitACommitObject(stream *bytes.Reader) (*ObjectCommit, error) {
	_, err := zlibDeflate(stream)
	if err != nil {
		return nil, fmt.Errorf("Error on emit a commit object %v", err)
	}
	// dataStr := string(data[:10])
	// fmt.Println(dataStr)
	// fmt.Println(len(data), stream.Size()-int64(stream.Len()))

	return &ObjectCommit{}, nil
}

type Pack struct {
	Packreader *bytes.Reader
	size       uint32
	version    uint32
	hash       string
	Objects    []Object
}

func NewPack(packBytes []byte) *Pack {
	packBytes = packBytes[4:]

	versionNum := binary.BigEndian.Uint32(packBytes[:4])
	totalSize := binary.BigEndian.Uint32(packBytes[4 : 4+4])

	packLen := len(packBytes)
	hashSize := 20
	hash := hex.EncodeToString(packBytes[packLen-hashSize:])
	packBytes = packBytes[8 : packLen-hashSize] // omit version, size and hash

	return &Pack{
		Packreader: bytes.NewReader(packBytes),
		size:       totalSize,
		version:    versionNum,
		hash:       hash,
	}
}

func (pck *Pack) ExtractObjects() error {
	for pck.Packreader.Len() > 0 {
		_, err := EmitAGitObject(pck.Packreader)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}
