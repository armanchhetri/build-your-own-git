package main

import (
	"fmt"
	// Uncomment this block to pass the first stage!
	// "flag"
	"os"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	//
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}
	subcommand, err := NewSubCommand(os.Args[1], os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	err = subcommand.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\nUsage: %v\n", err, subcommand.Usage())
		os.Exit(1)
	}

	// pprint := flag.Bool("p", true, "Pretty print the file contents")
	// objType := flag.Bool("t", false, "Only print the object type")
	// objSize := flag.Bool("s", false, "Only print the object size")
	// exitWith0 := flag.Bool("e", false, "Exit with status code 0 is file is not malformed")
	// flag.Parse()

	// objName := os.Args[1]
	// fmt.Printf("%v, %v, %v,%v, %v\n", objName, *pprint, *objType, *objSize, *exitWith0)

}
