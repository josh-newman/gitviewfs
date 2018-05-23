package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"log"
	"os"
	"path"
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

			fmt.Printf("  %s|- %s\n", dirPadding, getDisplayName(&entry, branchTree))
		}
	}
}

func getDisplayName(entry *object.TreeEntry, tree *object.Tree) string {
	switch entry.Mode {
	case filemode.Empty:
		return fmt.Sprintf("%s [0]", entry.Name)
	case filemode.Dir:
		return fmt.Sprintf("%s/", entry.Name)
	case filemode.Regular, filemode.Deprecated:
		file, err := tree.TreeEntryFile(entry)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "get file failed for entry %s", entry.Hash))
		}
		var preview string
		if isBinary, err := file.IsBinary(); err != nil {
			log.Fatalf("Unable to test if file %s is binary: %s", file.Hash, err)
		} else if isBinary {
			preview = "(binary)"
		} else {
			reader, err := file.Reader()
			defer reader.Close()
			if err != nil {
				log.Fatalf("Unable to read file %s: %s", file.Hash, err)
			}
			fileHead, err := readBoundedString(bufio.NewReader(reader), 40)
			if err != nil {
				log.Fatalf("Unable to read head of file %s: %s", file.Hash, err)
			}
			preview = fmt.Sprintf(`"%s"`, fileHead)
		}
		return fmt.Sprintf("%s [%s] %s", file.Name, file.Hash.String()[:8], preview)
	case filemode.Executable:
		return fmt.Sprintf("%s [* %s]", entry.Name, entry.Hash.String()[:8])
	case filemode.Symlink:
		return fmt.Sprintf("%s [->]", entry.Name)
	case filemode.Submodule:
		return fmt.Sprintf("%s [@]", entry.Name)
	}

	log.Fatalf("Unrecognized mode %s for entry: %s", entry.Mode, entry.Hash)
	panic("unreachable")
}

func readBoundedString(reader *bufio.Reader, maxLen int) (string, error) {
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
