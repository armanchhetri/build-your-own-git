package treecommit

import (
	"flag"
	"fmt"

	"github.com/codecrafters-io/git-starter-go/internal/treewriter"
)

type Treecommit struct {
	Fs            *flag.FlagSet
	parentHash    string
	currentHash   string
	commitMessage string
}

func (t *Treecommit) Initialize(args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("Not enough arguments")
	}
	t.currentHash = args[0]
	t.Fs.StringVar(&t.parentHash, "p", "", "Parent tree hash")
	t.Fs.StringVar(&t.commitMessage, "m", "", "Commit Message")
	t.Fs.Parse(args[1:])
	return nil
}

func (t *Treecommit) Usage() string {
	return "Write a tree object"
}

func (t *Treecommit) Run() error {
	commitData := `tree ` + t.currentHash + `
parent ` + t.parentHash + `
author Arman Chhetri <armanchhetri44@gmail.com> 1712252028 +0545
commiter Arman Chhetri <armanchhetri44@gmail.com> 1712252028 +0545

` + t.commitMessage + "\n"
	header := fmt.Sprintf("commit %d\x00", len(commitData))
	commitDataBytes := []byte(header + commitData)
	hash, err := treewriter.WriteObject(commitDataBytes)
	if err != nil {
		return err
	}
	// fmt.Println(commitData)
	fmt.Printf("%x\n", hash)
	return nil
}
