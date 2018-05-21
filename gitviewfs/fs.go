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
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"github.com/josh-newman/git-view-fs/gitviewfs/fserror"
	"github.com/hanwen/go-fuse/fuse/nodefs"
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
	node, ferr := f.findNode(name)
	if ferr != nil {
		if ferr.UnexpectedErr != nil {
			f.logger.Printf("unexpected error: %s", ferr.UnexpectedErr)
		}
		return nil, ferr.Status
	}

	var attr fuse.Attr
	switch n := node.(type) {
	case fstree.DirNode:
		attr.Mode = fuse.S_IFDIR | 0555
	case fstree.FileNode:
		attr.Mode = fuse.S_IFREG | 0444
		if n.File().Mode == filemode.Executable {
			attr.Mode |= 0555
		}
		attr.Size = uint64(n.File().Size)
	default:
		f.logger.Printf("skipping node: %v", node)
		return nil, fuse.ENOENT
	}

	return &attr, fuse.OK
}

func (f *gitviewfs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	node, ferr := f.findNode(name)
	if ferr != nil {
		if ferr.UnexpectedErr != nil {
			f.logger.Printf("unexpected error: %s", ferr.UnexpectedErr)
		}
		return nil, ferr.Status
	}

	dirNode, ok := node.(fstree.DirNode)
	if !ok {
		return nil, fuse.ENOTDIR
	}

	children, ferr := dirNode.Children()
	if ferr != nil {
		if ferr.UnexpectedErr != nil {
			f.logger.Printf("unexpected error: %s", ferr.UnexpectedErr)
		}
		return nil, ferr.Status
	}

	var entries []fuse.DirEntry
	for name, child := range children {
		entry := fuse.DirEntry{Name: name}
		switch n := child.(type) {
		case fstree.DirNode:
			entry.Mode = fuse.S_IFDIR | 0555
		case fstree.FileNode:
			entry.Mode = fuse.S_IFREG | 0444
			if n.File().Mode == filemode.Executable {
				entry.Mode |= 0555
			}
		default:
			f.logger.Printf("skipping child: %v", node)
		}
		entries = append(entries, entry)
	}

	return entries, fuse.OK
}

func (f *gitviewfs) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	node, ferr := f.findNode(name)
	if ferr != nil {
		if ferr.UnexpectedErr != nil {
			f.logger.Printf("unexpected error: %s", ferr.UnexpectedErr)
		}
		return nil, ferr.Status
	}

	fileNode, ok := node.(fstree.FileNode)
	if !ok {
		return nil, fuse.EINVAL
	}

	// TODO(josh-newman): Read efficiently instead of copying the whole file into memory.
	reader, err := fileNode.File().Reader()
	if err != nil {
		f.logger.Printf("error creating file reader: %s", err)
		return nil, fuse.EIO
	}
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		f.logger.Printf("error reading file: %s", err)
		return nil, fuse.EIO
	}

	return nodefs.NewDataFile(bytes), fuse.OK
}

func (f *gitviewfs) findNode(name string) (fstree.Node, *fserror.Error) {
	node := f.fstree
	if name == "" {
		return node, nil
	}
	for _, part := range strings.Split(name, "/") {
		if dirNode, ok := node.(fstree.DirNode); ok {
			children, ferr := dirNode.Children()
			if ferr != nil {
				return nil, ferr
			}

			if child, ok := children[part]; !ok {
				return nil, fserror.Expected(fuse.ENOENT)
			} else {
				node = child
			}
		} else {
			return nil, fserror.Expected(fuse.ENOENT)
		}
	}
	return node, nil
}
