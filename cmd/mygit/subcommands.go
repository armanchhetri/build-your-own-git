package main

import (
	"bufio"
	"compress/zlib"
	"flag"
	"fmt"

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
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	// extract header and size
	// Looks like doing a lot to do simple thing
	var fileType []byte
	var size []byte
	startedSize := false
	for scanner.Scan() {
		scanByte := scanner.Bytes()[0]
		if scanByte == 0 {
			break
		}
		if scanByte == ' ' {
			startedSize = true
			continue
		}
		if startedSize {
			startedSize = true
			size = append(size, scanByte)
		} else {
			fileType = append(fileType, scanByte)
		}
	}
	if c.objType {
		fmt.Print(string(fileType))
		return nil
	}
	if c.objSize {
		fmt.Print(string(size))
		return nil
	}
	if c.exitWith0 {
		os.Exit(0)
	}

	// print data
	for scanner.Scan() {
		fmt.Print(scanner.Text())
	}

	r.Close()

	return nil
}
func (c *Catfile) Usage() string {
	return "git cat-file : Prints the contents of a git object"
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

	default:
		return nil, fmt.Errorf("Unknown command %s\nUsage: git <command> <args>", subComName)
	}
}
