package tree

import (
	"bufio"
	"bytes"

	// "bytes"
	"compress/zlib"
	"flag"
	"fmt"

	// "io"
	"os"
	"path/filepath"
	"strings"
)

// type Tree struct {
// 	shasum     string
// 	objectType string
// 	objectName string
// 	mode       string
// }

type LsTree struct {
	Fs       *flag.FlagSet
	ObjName  string
	NameOnly bool
	// entry    []Tree
}

func (lstree *LsTree) Initialize(args []string) error {
	lstree.Fs.BoolVar(&lstree.NameOnly, "name-only", false, "Name only")
	err := lstree.Fs.Parse(args)

	if err != nil {
		return err
	}
	lstree.ObjName = args[len(args)-1]
	// fmt.Println(lstree.NameOnly)
	return nil
}

func (lstree *LsTree) Usage() string {
	return "Print the directory structure of the given tree object"
}

func (lstree *LsTree) Run() error {
	objPath := filepath.Join(".git/objects", lstree.ObjName[:2], lstree.ObjName[2:])
	objFile, err := os.Open(objPath)
	if err != nil {
		return err
	}
	defer objFile.Close()
	zlibReader, err := zlib.NewReader(objFile)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(zlibReader)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		scanByte := scanner.Bytes()[0]
		if scanByte == 0 {
			break
		}
	}
	// start of the file content
	var accumulator bytes.Buffer
	for scanner.Scan() {

		scanBytes := scanner.Bytes()
		// fmt.Printf("%c\n", scanBytes[0])
		scanByte := scanBytes[0]
		if scanByte == 0 {
			hash := make([]byte, 20)
			i := 0
			for {
				if i == 20 {
					break
				}
				scanner.Scan()
				scanByte = scanner.Bytes()[0]
				hash[i] = scanByte

				i++
			}
			hashHex := fmt.Sprintf("%x", hash)

			modeName := strings.SplitN(accumulator.String(), " ", 2)
			mode := modeName[0]

			Name := modeName[1]
			entityType := "blob"
			if mode[0] != '1' {
				entityType = "tree"
				mode = "0" + mode
			}
			if lstree.NameOnly {
				fmt.Println(Name)
			} else {
				fmt.Printf("%s %s %s\t%s\n", mode, entityType, hashHex, Name)
			}
			accumulator.Reset()
			continue
		}
		accumulator.WriteByte(scanByte)

	}

	return nil

}
