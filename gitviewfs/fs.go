package gitviewfs

import (
	"gopkg.in/src-d/go-git.v4"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/fuse"
	"log"
	"strings"
	"io/ioutil"
	"os"
	"github.com/josh-newman/git-view-fs/gitviewfs/gitfstree"
	"github.com/josh-newman/git-view-fs/gitviewfs/fstree"
)

type gitviewfs struct {
	pathfs.FileSystem
	fstree fstree.Node
	logger *log.Logger
}

func New(repo *git.Repository) (pathfs.FileSystem, error) {
	tree, err := gitfstree.New(repo)
	if err != nil {
		return nil, err
	}

	return &gitviewfs{
		FileSystem: pathfs.NewDefaultFileSystem(),
		fstree: tree,
		logger: log.New(ioutil.Discard, "gitviewfs", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC),
	}, nil
}

func (f *gitviewfs) String() string {
	// TODO(josh-newman): Add repository path.
	return "gitviewfs"
}

func (f *gitviewfs) SetDebug(debug bool) {
	if debug {
		f.logger.SetOutput(os.Stderr)
	} else {
		f.logger.SetOutput(ioutil.Discard)
	}
}

func (f *gitviewfs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	node := f.fstree
	for _, part := range strings.Split(name, "/") {
		if dirNode, ok := node.(fstree.DirNode); ok {
			children, ferr := dirNode.Children()
			if ferr != nil {
				f.logger.Printf("unexpected error: %s", ferr.UnexpectedErr)
				return nil, ferr.Status
			}

			if child, ok := children[part]; !ok {
				return nil, fuse.ENOENT
			} else {
				node = child
			}
		} else {
			return nil, fuse.ENOENT
		}
	}

	var attr fuse.Attr
	switch n := node.(type) {
	case fstree.DirNode:
		attr.Mode = fuse.S_IFDIR | 0555
	case fstree.FileNode:
		attr.Mode = fuse.S_IFREG | 0444
		if n.Executable() {
			attr.Mode |= 0111
		}
		attr.Size = uint64(n.Size())
	default:
		log.Printf("skipping node: %v", node)
		return nil, fuse.ENOENT
	}

	return &attr, fuse.OK
}
