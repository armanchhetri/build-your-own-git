package treewriter

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Treewriter struct {
	Fs *flag.FlagSet
}

func NewTreewriter() *Treewriter {
	return &Treewriter{}
}

func (t *Treewriter) Initialize(args []string) error {
	return nil
}

func (t *Treewriter) Usage() string {
	return "Write all the objects to .git/objects directory"
}

func (t *Treewriter) Run() error {
	// get current directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	hash, err := createAndWriteObjects(currentDirectory)
	if err != nil {
		return err
	}
	hash = hash[len(hash)-20:]
	fmt.Printf("%x", hash)
	return nil
}

func createAndWriteObjects(path string) ([]byte, error) {
	if strings.Contains(path, ".git") {
		return []byte{}, nil
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var treeContents []byte
	for _, file := range files {
		if file.IsDir() {
			treeContent, err := createAndWriteObjects(path + "/" + file.Name())
			if err != nil {
				return nil, err
			}
			treeContents = append(treeContents, treeContent...)
		} else {
			treeContent, _ := createBlobObject(path, file.Name())
			treeContents = append(treeContents, treeContent...)
		}
	}
	treeHdrStr := fmt.Sprintf("tree %d\x00", len(treeContents))
	treeContents = append([]byte(treeHdrStr), treeContents...)
	hash, err := writeObject(treeContents)
	if err != nil {
		return nil, err
	}
	treeContentStr := fmt.Sprintf("40000 %s\x00", filepath.Base(path)) // 0 is prepended by ls-tree and git's cat-file
	treeContent := append([]byte(treeContentStr), hash[:]...)
	// treeContent = append(treeContent, 0)
	return treeContent, nil
}

func createBlobObject(path string, filename string) ([]byte, error) {
	data, _ := os.ReadFile(filepath.Join(path, filename))

	header := fmt.Sprintf("blob %d\x00", len(data))
	finalData := append([]byte(header), data...)
	hash, err := writeObject(finalData)
	if err != nil {
		return nil, err
	}

	treeContentStr := fmt.Sprintf("100644 %s\x00", filename)
	treeContent := append([]byte(treeContentStr), hash[:]...)
	// treeContent = append(treeContent, 0)
	return treeContent, nil
}

func writeObject(data []byte) ([]byte, error) {
	hash := sha1.Sum(data)
	hashStr := fmt.Sprintf("%x", hash)
	prefix := hashStr[:2]
	objFilename := hashStr[2:]
	objDir := ".git/objects"
	err := os.MkdirAll(filepath.Join(objDir, prefix), 0o755)
	if err != nil {
		return nil, err
	}
	var compressedData bytes.Buffer
	w := zlib.NewWriter(&compressedData)
	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}

	w.Close()
	err = os.WriteFile(filepath.Join(objDir, prefix, objFilename), compressedData.Bytes(), 0o644)
	if err != nil {
		return nil, err
	}
	return hash[:], nil
}
