package main

import (
	"flag"
	"fmt"

	"github.com/codecrafters-io/git-starter-go/internal/clone"
	"github.com/codecrafters-io/git-starter-go/internal/general"
	"github.com/codecrafters-io/git-starter-go/internal/tree"
	"github.com/codecrafters-io/git-starter-go/internal/treecommit"
	"github.com/codecrafters-io/git-starter-go/internal/treewriter"
)

type Subcommand interface {
	Run() error
	Initialize(args []string) error
	Usage() string
}

func NewSubCommand(subComName string, args []string) (Subcommand, error) {
	switch subComName {
	case "init":
		initer := &general.Init{
			Fs: flag.NewFlagSet("init", flag.ExitOnError),
		}
		if initer.Initialize(args) != nil {
			return nil, fmt.Errorf("Error initializing command")
		}
		return initer, nil
	case "cat-file":
		catter := &general.Catfile{
			Fs: flag.NewFlagSet("cat-file", flag.ExitOnError),
		}
		catter.Initialize(args[1:])
		return catter, nil

	case "hash-object":
		hasher := &general.HashObject{
			Fs: flag.NewFlagSet("hash-object", flag.ExitOnError),
		}
		hasher.Initialize(args[1:])
		return hasher, nil

	case "ls-tree":
		lstreer := &tree.LsTree{Fs: flag.NewFlagSet("ls-tree", flag.ExitOnError)}
		err := lstreer.Initialize(args[1:])
		if err != nil {
			return lstreer, err
		}
		return lstreer, nil

	case "write-tree":
		writer := &treewriter.Treewriter{Fs: flag.NewFlagSet("write-tree", flag.ExitOnError)}
		writer.Initialize(args[1:])
		return writer, nil

	case "commit-tree":
		commiter := &treecommit.Treecommit{Fs: flag.NewFlagSet("commit-tree", flag.ExitOnError)}
		commiter.Initialize(args[1:])
		return commiter, nil

	case "clone":
		cloner := &clone.Clone{Fs: flag.NewFlagSet("clone", flag.ExitOnError)}
		cloner.Initialize(args[1:])
		return cloner, nil
	default:
		return nil, fmt.Errorf("Unknown command %s\nUsage: git <command> <args>", subComName)
	}
}
