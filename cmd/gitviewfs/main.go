package main

import (
	"log"
	"gopkg.in/src-d/go-git.v4"
	"github.com/pkg/errors"
	"github.com/josh-newman/git-view-fs/gitviewfs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse"
	"flag"
)

var debug = flag.Bool("debug", false, "enable debug logging")

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatalf("Expected two arguments: /mount/point /path/to/git/repository")
	}
	mountPath := flag.Arg(0)
	repoPath := flag.Arg(1)

	repo, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		log.Fatalf("No git repository found: %s", repoPath)
	} else if err != nil {
		log.Fatal(errors.Wrap(err, "open git repository failed"))
	}

	gfs, err := gitviewfs.New(repo)
	if err != nil {
		log.Fatal(errors.Wrap(err, "create gitviewfs failed"))
	}

	nfs := pathfs.NewPathNodeFs(gfs, &pathfs.PathNodeFsOptions{Debug: *debug})
	connector := nodefs.NewFileSystemConnector(nfs.Root(), &nodefs.Options{Debug: *debug})
	server, err := fuse.NewServer(
		connector.RawFS(),
		mountPath,
		&fuse.MountOptions{
			FsName: "git:" + repoPath,
			Name: "gitviewfs",
			Debug: *debug,
		},
	)

	if err != nil {
		log.Fatalf("serve failed: %s", err)
	}

	server.Serve()
}
