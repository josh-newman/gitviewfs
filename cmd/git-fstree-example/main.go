package main

import (
	"bufio"
	"fmt"
	"github.com/josh-newman/git-view-fs/gitviewfs/fstree"
	"github.com/josh-newman/git-view-fs/gitviewfs/gitfstree"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
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
			ferr := printChildren(child, depth+1)
			if ferr != nil {
				return ferr
			}
		case fstree.FileNode:
			file := child.File()
			var preview string
			if isBinary, err := file.IsBinary(); err != nil {
				log.Fatalf("Unable to test if file %s is binary: %s", file.Hash, err)
			} else if isBinary {
				preview = "(binary)"
			} else {
				fileHead, err := readBoundedString(file, 40)
				if err != nil {
					log.Fatalf("Unable to read head of file %s: %s", file.Hash, err)
				}
				preview = fmt.Sprintf(`"%s"`, fileHead)
			}
			executableSuffix := ""
			if file.Mode == filemode.Executable {
				executableSuffix = "*"
			}
			fmt.Printf("%s%s%s [%d] %s\n", indent, name, executableSuffix, child.File().Size, preview)
		}
	}
	return nil
}

func readBoundedString(file *object.File, maxLen int) (string, error) {
	rawReader, err := file.Reader()
	if err != nil {
		return "", err
	}
	defer rawReader.Close()
	reader := bufio.NewReader(rawReader)
	var runes []rune
	for len(runes) < maxLen+1 {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		if !unicode.IsPrint(r) {
			if unicode.IsSpace(r) {
				r = ' '
			} else {
				r = unicode.ReplacementChar
			}
		}
		runes = append(runes, r)
	}
	if len(runes) == maxLen+1 {
		// The file is longer than maxLen, so we need to abridge.
		runes = runes[:maxLen]
		runes[maxLen-1] = '…'
	}
	return string(runes), nil
}
