package main

import (
	"os"
	"log"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"github.com/josh-newman/git-view-fs/gitviewfs/gitfstree"
	"github.com/josh-newman/git-view-fs/gitviewfs/fstree"
	"sort"
	"strings"
	"fmt"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Expected exactly one argument: /path/to/git/repository")
	}
	repoPath := os.Args[1]

	repo, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		log.Fatalf("No git repository found: %s", repoPath)
	} else if err != nil {
		log.Fatal(errors.Wrap(err, "open git repository failed"))
	}

	tree, err := gitfstree.New(repo)
	if err != nil {
		log.Fatalf("error creating gitfstree: %s", err)
	}

	err = printChildren(tree.(fstree.DirNode), 0)
	if err != nil {
		log.Fatalf("error printing: %s", err)
	}
}

func printChildren(node fstree.DirNode, depth int) error {
	children, ferr := node.Children()
	if ferr != nil {
		return ferr
	}

	var childNames []string
	for name, _ := range children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	indent := strings.Repeat("  ", depth)
	for _, name := range childNames {
		switch child := children[name].(type) {
		case fstree.DirNode:
			fmt.Printf("%s%s/\n", indent, name)
			ferr := printChildren(child, depth + 1)
			if ferr != nil {
				return ferr
			}
		case fstree.FileNode:
			executableSuffix := ""
			if child.Executable() {
				executableSuffix = "*"
			}
			fmt.Printf("%s%s%s [%d]\n", indent, name, executableSuffix, child.Size())
		}
	}
	return nil
}
