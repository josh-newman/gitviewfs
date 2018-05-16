package main

import (
	"gopkg.in/src-d/go-git.v4"
	"log"
	"github.com/pkg/errors"
	"io"
	"fmt"
	"os"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"strings"
	"path"
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

	branchRefs, err := repo.Branches()
	if err != nil {
		log.Fatal(errors.Wrap(err, "list branches failed"))
	}

	for {
		branchRef, err := branchRefs.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(errors.Wrap(err, "next branch failed"))
		}

		branchCommit, err := repo.CommitObject(branchRef.Hash())
		if err == plumbing.ErrObjectNotFound {
			log.Fatalf("Branch %s points to invalid or non-commit ref %s", branchRef.Name(), branchRef.Hash())
		} else if err != nil {
			log.Fatal(errors.Wrap(err, "find branch commit failed"))
		}

		branchTree, err := branchCommit.Tree()
		if err != nil {
			log.Fatal(errors.Wrap(err, "find branch tree failed"))
		}

		walker := object.NewTreeWalker(branchTree, true, map[plumbing.Hash]bool{})
		fmt.Println(branchRef.Name())
		for {
			entryPath, entry, err := walker.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(errors.Wrap(err, "tree walking failed"))
			}

			dirPath, _ := path.Split(entryPath)
			var dirDepth int
			if len(dirPath) > 0 {
				dirDepth = len(strings.Split(path.Clean(dirPath), "/"))
			}
			dirPadding := strings.Repeat("|  ", dirDepth)
			fmt.Printf("  %s|- %s\n", dirPadding, getDisplayName(entry))
		}
	}
}

func getDisplayName(entry object.TreeEntry) string {
	switch entry.Mode {
	case filemode.Empty:
		return fmt.Sprintf("%s [0]", entry.Name)
	case filemode.Dir:
		return fmt.Sprintf("%s/", entry.Name)
	case filemode.Regular, filemode.Deprecated:
		return fmt.Sprintf("%s [%s]", entry.Name, entry.Hash.String())
	case filemode.Executable:
		return fmt.Sprintf("%s [* %s]", entry.Name, entry.Hash.String())
	case filemode.Symlink:
		return fmt.Sprintf("%s [->]", entry.Name)
	case filemode.Submodule:
		return fmt.Sprintf("%s [@]", entry.Name)
	}

	log.Fatalf("Unrecognized mode %s for tree entry: %s", entry.Mode, entry.Hash)
	panic("unreachable")
}
