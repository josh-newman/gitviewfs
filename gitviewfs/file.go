package gitviewfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/josh-newman/gitviewfs/gitviewfs/fstree"
	"io"
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

	// Seek to requested offset by reading and discarding.
	// TODO(josh-newman): Seek more efficiently? Avoid O(n^2) reading across multiple requests.
	for off > 0 {
		nDiscarded, err := io.CopyN(ioutil.Discard, reader, off)
		off -= nDiscarded
		if err == io.EOF {
			off = 0
		} else if err != nil {
			f.logger.Printf("error seeking in file: %s", err)
			return nil, fuse.EIO
		}
	}

	nRead, err := reader.Read(dest)
	dest = dest[:nRead]
	if err != nil {
		f.logger.Printf("error reading file: %s", err)
		return nil, fuse.EIO
	}

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
