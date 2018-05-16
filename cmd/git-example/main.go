package main

import (
	"gopkg.in/src-d/go-git.v4"
	"log"
	"github.com/pkg/errors"
	"io"
	"fmt"
)

func main() {
	repoPath := "/tmp/git-test"

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

		fmt.Println(branchRef)
	}
}
