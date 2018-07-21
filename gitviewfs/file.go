package gitviewfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/josh-newman/gitviewfs/gitviewfs/fstree"
	"io/ioutil"
	"log"
)

type file struct {
	node   fstree.FileNode
	logger *log.Logger
	nodefs.File
}

func newFile(node fstree.FileNode, logger *log.Logger) nodefs.File {
	return nodefs.NewReadOnlyFile(&file{
		node:   node,
		logger: logger,
		File:   nodefs.NewDefaultFile(),
	})
}

func (f *file) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	reader, err := f.node.File().Reader()
	if err != nil {
		f.logger.Printf("error creating file reader: %s", err)
		return nil, fuse.EIO
	}
	defer reader.Close()

	// TODO(josh-newman): Read efficiently instead of copying the whole file into memory.
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		f.logger.Printf("error reading file: %s", err)
		return nil, fuse.EIO
	}

	_ = copy(dest, bytes)
	return fuse.ReadResultData(dest), fuse.OK
}

func (f *file) GetAttr(out *fuse.Attr) fuse.Status {
	if mode := computeFuseFileMode(f.node.File().Mode); mode != 0 {
		out.Mode = mode
	} else {
		f.logger.Printf("skipping file child: %v", f.node)
		return fuse.ENOENT
	}
	out.Size = uint64(f.node.File().Size)
	return fuse.OK
}
