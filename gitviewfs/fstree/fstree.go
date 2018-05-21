package fstree

import (
	"github.com/josh-newman/git-view-fs/gitviewfs/fserror"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Node interface{}

type DirNode interface {
	Node
	Children() (map[string]Node, *fserror.Error)
}

type FileNode interface {
	Node
	File() *object.File
}
