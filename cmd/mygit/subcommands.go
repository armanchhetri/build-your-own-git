package main

import (
	// "bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"

	// "strconv"
	"strings"

	// "io"
	"os"
	"path/filepath"
)

type Subcommand interface {
	Run() error
	Initialize(args []string) error
	Usage() string
}

type Init struct {
	fs *flag.FlagSet
}

func (in *Init) Initialize(args []string) error {

	return nil
}

func (in *Init) Usage() string {
	return "git init : Initializes an empty git repository in the current directory"
}

func (in *Init) Run() error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
	}

	fmt.Println("Initialized git directory")
	return nil
}

type Catfile struct {
	fs        *flag.FlagSet
	objName   string
	pprint    bool
	objType   bool
	objSize   bool
	exitWith0 bool
}

func (c *Catfile) Initialize(args []string) error {
	c.fs.BoolVar(&c.pprint, "p", true, "Pretty print")
	c.fs.BoolVar(&c.objType, "t", false, "Object type")
	c.fs.BoolVar(&c.objSize, "s", false, "Object size")
	c.fs.BoolVar(&c.exitWith0, "e", false, "Exit with 0")
	err := c.fs.Parse(args)
	if err != nil {
		return err
	}
	c.objName = args[len(args)-1]
	// if c.objName starts with - return error
	if c.objName[0] == '-' {
		return fmt.Errorf("Invalid object name")
	}
	return nil

}

func (c *Catfile) Run() error {
	filepath := filepath.Join(".git/objects", c.objName[:2], c.objName[2:])

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	r, err := zlib.NewReader(file)
	if err != nil {
		return err
	}
	data, _ := io.ReadAll(r)
	dataStr := string(data)
	objTypeLen := strings.IndexByte(dataStr, ' ')
	sizeLen := strings.IndexByte(dataStr, 0)
	objType := dataStr[:objTypeLen]
	size := dataStr[objTypeLen+1 : sizeLen]
	body := dataStr[sizeLen+1:]
	if c.objType {
		fmt.Print(string(objType))
		return nil
	}
	if c.objSize {
		fmt.Print(string(size))
		return nil
	}
	if c.exitWith0 {
		os.Exit(0)
	}
	fmt.Print(body)
	r.Close()

	return nil

	// short version
	// var decompressed bytes.Buffer
	// decompressed.ReadFrom(r)
	// parts := bytes.SplitN(decompressed.Bytes(), []byte{'\x00'}, 2)

	// Long implementation may come handy for future reference
	// scanner := bufio.NewScanner(r)
	// scanner.Split(bufio.ScanBytes)
	// // extract header and size
	// // Looks like doing a lot to do simple thing
	// var fileType []byte
	// var size []byte
	// startedSize := false
	// for scanner.Scan() {
	// 	scanByte := scanner.Bytes()[0]
	// 	if scanByte == 0 {
	// 		break
	// 	}
	// 	if scanByte == ' ' {
	// 		startedSize = true
	// 		continue
	// 	}
	// 	if startedSize {
	// 		startedSize = true
	// 		size = append(size, scanByte)
	// 	} else {
	// 		fileType = append(fileType, scanByte)
	// 	}
	// }

	// // print data
	// for scanner.Scan() {
	// 	fmt.Print(scanner.Text())
	// }

}
func (c *Catfile) Usage() string {
	return "git cat-file : Prints the contents of a git object"
}

type HashObject struct {
	fs          *flag.FlagSet
	writeObject bool // if true write the object's output to .git/objects/<2char>/<remining char>
	objName     string
}

func (h *HashObject) Initialize(args []string) error {
	h.fs.BoolVar(&h.writeObject, "w", false, "write the object to the objects directory")
	h.fs.Parse(args)
	h.objName = args[len(args)-1]
	return nil
}

func (h *HashObject) Run() error {
	objDir := ".git/objects"
	_, err := os.Stat(objDir)
	if err != nil {
		return fmt.Errorf("Could not find valid git repository, Did you git init?")
	}
	fileInfo, err := os.Stat(h.objName)
	if err != nil {
		return fmt.Errorf("Could not find the file: %s \n", h.objName)
	}

	// read the file
	data, err := os.ReadFile(h.objName)
	if err != nil {
		return err
	}

	// header := []byte("blob" + strconv.Itoa(int(fileInfo.Size())) + string(0)) my version :)
	headerStr := fmt.Sprintf("blob %d%c", fileInfo.Size(), 0)
	header := []byte(headerStr)
	finalData := append(header, data...)
	hash := sha1.Sum(finalData)
	hashStr := fmt.Sprintf("%x", hash)
	if h.writeObject {
		prefix := hashStr[:2]
		filename := hashStr[2:]

		err = os.Mkdir(filepath.Join(objDir, prefix), 0755)
		if err != nil {
			return err
		}
		var compressedData bytes.Buffer
		w := zlib.NewWriter(&compressedData)
		_, err = w.Write(finalData)
		if err != nil {
			return err
		}

		w.Close()
		err = os.WriteFile(filepath.Join(objDir, prefix, filename), compressedData.Bytes(), 0744)

		if err != nil {
			return err
		}
	}

	fmt.Print(hashStr)

	return nil
}

func (h *HashObject) Usage() string {
	return "git hash-object : Computes the SHA1 hash of an object"
}

func NewSubCommand(subComName string, args []string) (Subcommand, error) {
	switch subComName {
	case "init":
		initer := &Init{
			fs: flag.NewFlagSet("init", flag.ExitOnError),
		}
		if initer.Initialize(args) != nil {
			return nil, fmt.Errorf("Error initializing command")
		}
		return initer, nil
	case "cat-file":
		catter := &Catfile{
			fs: flag.NewFlagSet("cat-file", flag.ExitOnError),
		}
		catter.Initialize(args[1:])
		return catter, nil

	case "hash-object":
		hasher := &HashObject{
			fs: flag.NewFlagSet("hash-object", flag.ExitOnError),
		}
		hasher.Initialize(args[1:])
		return hasher, nil

	default:
		return nil, fmt.Errorf("Unknown command %s\nUsage: git <command> <args>", subComName)
	}
}
