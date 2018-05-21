package fstree

import "github.com/josh-newman/git-view-fs/gitviewfs/fserror"

type Node interface{}

type DirNode interface {
	Node
	Children() (map[string]Node, *fserror.Error)
}

type FileNode interface {
	Node
	Size() uint64
	Executable() bool
}
