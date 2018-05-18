package main

import (
	"os"
	"log"
	"gopkg.in/src-d/go-git.v4"
	"github.com/pkg/errors"
	"github.com/josh-newman/git-view-fs/gitviewfs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Expected exactly two arguments: %s /mount/point /path/to/git/repository", os.Args[0])
	}
	mountPath := os.Args[1]
	repoPath := os.Args[2]

	repo, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		log.Fatalf("No git repository found: %s", repoPath)
	} else if err != nil {
		log.Fatal(errors.Wrap(err, "open git repository failed"))
	}

	nfs := pathfs.NewPathNodeFs(gitviewfs.New(repo), &pathfs.PathNodeFsOptions{Debug: true})
	connector := nodefs.NewFileSystemConnector(nfs.Root(), &nodefs.Options{Debug: true})
	server, err := fuse.NewServer(
		connector.RawFS(),
		mountPath,
		&fuse.MountOptions{
			FsName: "git:" + repoPath,
			Name: "gitviewfs",
			Debug: true,
		},
	)

	if err != nil {
		log.Fatalf("serve failed: %s", err)
	}

	server.Serve()
}
