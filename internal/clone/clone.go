package clone

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Clone struct {
	Fs  *flag.FlagSet
	URL string
}

func (t *Clone) Initialize(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Not enough arguments")
	}
	t.URL = args[len(args)-1]
	return nil
}

func (t *Clone) Usage() string {
	return "Clone a repository: git clone <url>"
}

// Sends GET request to the reference url and extracts references
// URL: root url to the repo
func getReferences(URL string) ([]string, error) {
	referencePostfix := "/info/refs?service=git-upload-pack" // defined by the git http smart protocol
	referenceURL := URL + referencePostfix

	resp, err := http.Get(referenceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not clone the repository %v", URL)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	refs, err := extractRefs(data)
	if err != nil {
		return nil, err
	}

	return refs, nil
}

// masterHash: Hash of the commit object that points to Master. This is where data transfer starts
// sends post request to the endpoint /git-upload-pack to get actual packfile
func getPackFile(URL, masterHash string) ([]byte, error) {
	packPostfix := "/git-upload-pack"
	objURL := URL + packPostfix
	initiateProtoMsg := fmt.Sprintf("0032want %v\n00000009done\n", masterHash)
	reqBody := bytes.NewBufferString(initiateProtoMsg)
	contentType := "application/x-git-upload-pack-request"

	resp, err := http.Post(objURL, contentType, reqBody)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error on fetching pack files %v", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// first 8 bytes are ack things(https://www.git-scm.com/docs/gitprotocol-pack#:~:text=0009done%5Cn%0A%0A%20%20%20S%3A-,0008NAK%5Cn,-S%3A%20%5BPACKFILE%5D)

	return data[8:], nil
}

func verifyPackfile(packfile []byte) bool {
	checksumLen := 20
	packOnly := packfile[:len(packfile)-checksumLen]
	checksum := packfile[len(packfile)-checksumLen:]
	expectedChecksum := sha1.Sum(packOnly)
	return bytes.Equal(expectedChecksum[:], checksum)
}

type ObjectType uint8

const (
	OBJ_COMMIT    ObjectType = 1
	OBJ_TREE      ObjectType = 2
	OBJ_BLOB      ObjectType = 3
	OBJ_TAG       ObjectType = 4
	OBJ_OFS_DELTA ObjectType = 6
	OBJ_REF_DELTA ObjectType = 7
)

var OJB_MAP = map[ObjectType]string{
	OBJ_COMMIT:    "OBJ_COMMIT",
	OBJ_TREE:      "OBJ_TREE",
	OBJ_BLOB:      "OBJ_BLOB",
	OBJ_TAG:       "OBJ_TAG",
	OBJ_OFS_DELTA: "OBJ_OFS_DELTA",
	OBJ_REF_DELTA: "OBJ_REF_DELTA",
}

type Object struct {
	objType string
	size    int
	data    []byte
}

func (t *Clone) Run() error {
	// get references
	refs, err := getReferences(t.URL)
	if err != nil {
		return err
	}

	masterHash, err := getMasterObjHash(refs)
	if err != nil {
		return err
	}

	packfile, err := getPackFile(t.URL, masterHash)
	if err != nil {
		return err
	}

	if ok := verifyPackfile(packfile); !ok {
		return errors.New("Could not verify the packfile")
	}
	fmt.Println(packfile[:20])
	// skip "pack"
	packfile = packfile[4:]

	versionNum := binary.BigEndian.Uint32(packfile[:4])
	totalSize := binary.BigEndian.Uint32(packfile[4 : 4+4])

	packfile = packfile[8:]

	fmt.Println(versionNum, totalSize, packfile[0:3])

	for {
		// scan objects
		index := 0
		firstByte := packfile[index]
		fmt.Println(firstByte, packfile[index+1])
		shouldReadNext := firstByte > 128 // MSB == 1
		var mask uint8 = 112              // 0b01110000
		objectType := firstByte & mask >> 4
		objectTypeName := OJB_MAP[ObjectType(objectType)]
		fmt.Println(objectTypeName, shouldReadNext)
		mask = 15 // 0b00001111
		objectSize := uint64(firstByte & mask)
		shift := 4
		// first 4 bits form least significant byte for integer: size  eg: 0000XXXX, first 4 are 0 and last 4 are actual bits
		// other 7 bits are prepended to first 4 while first bit is 1
		// for eg: byte array [144 15 120] will form integer bit sequence of 0b11110000 = 240

		for shouldReadNext {
			index++
			nextByte := packfile[index]
			shouldReadNext = nextByte > 128 // MSB == 1
			var msbMask uint8 = 127
			nextByte = nextByte & msbMask
			fmt.Println(objectSize, index)
			objectSize = objectSize | uint64(nextByte)<<shift
			shift += 7
			fmt.Println(objectSize, index)
		}

		object := packfile[index+1:]
		objstream := bytes.NewBuffer(object)

		r, err := zlib.NewReader(objstream)
		if err != nil {
			return err
		}
		data, _ := io.ReadAll(r)
		dataStr := string(data)
		fmt.Printf("%v \n len = %d\n", dataStr, len(data))

		fmt.Println()
		offset := len(object) - objstream.Len()
		if offset <= 0 {
			break
		}
		packfile = packfile[offset+1:]

	}

	fmt.Println("NOTE: Clone is WORK IN PROGRESS")
	return nil
}

func getMasterObjHash(refs []string) (string, error) {
	for _, ref := range refs {
		// hashValue := strings.SplitN(ref, " ", 2)
		// hash := hashValue[0]
		// value := hashValue[1]
		var hash, val string
		fmt.Sscanf(ref, "%v %v", &hash, &val) // saw a better way :)
		if val == "refs/heads/master" {
			return hash, nil
		}
	}
	return "", errors.New("Did not get refs/heads/master")
}

func extractRefs(data []byte) ([]string, error) {
	// should be of the form
	// smart_reply     =  PKT-LINE("# service=$servicename" LF)
	//  "0000"
	//  *1("version 1")
	//  ref_list
	//  "0000"
	// ref_list        =  empty_list / non_empty_list

	// skip first line
	lineSize, err := bToUint16(data[:4])
	if err != nil {
		return nil, err
	}
	skipLine4Zeros := lineSize + 4
	refList := make([]string, 0)
	currIdx := skipLine4Zeros
	// while not considering version
	for int(currIdx) < len(data)-4 { // 4 is subtracted to rule out trailing 4 zeros
		currData := data[currIdx:]
		size := currData[:4]
		sizeInt, err := bToUint16(size)
		if err != nil {
			return nil, err
		}
		refList = append(refList, string(currData[4:sizeInt]))
		currIdx += sizeInt
	}
	// remove caps from first entry
	refList[0] = strings.Split(refList[0], string(rune(0)))[0]
	return refList, nil
}

func bToUint16(bytes []byte) (uint16, error) {
	// fmt.Println(bytes)
	size := make([]byte, 2, 2)
	_, err := hex.Decode(size, bytes)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(size), nil
}
