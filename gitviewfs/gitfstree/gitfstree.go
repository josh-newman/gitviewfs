package gitfstree

import (
	"fmt"
	"github.com/josh-newman/git-view-fs/gitviewfs/fserror"
	"github.com/josh-newman/git-view-fs/gitviewfs/fstree"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"os"
	"strings"
)

func New(repo *git.Repository) (fstree.Node, error) {
	branchRefs, err := repo.Branches()
	if err != nil {
		return nil, errors.Wrap(err, "list branches failed")
	}

	node := branchesNode{repo: repo}
	for {
		branchRef, err := branchRefs.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fserror.Unexpected(errors.Wrap(err, "next branch failed"))
		}

		nameParts := strings.Split(string(branchRef.Name()), "/")
		node.entries = append(node.entries, branchesNodeEntry{nameParts: nameParts, branchRef: branchRef})
	}

	return &node, nil
}

type branchesNodeEntry struct {
	nameParts []string
	branchRef *plumbing.Reference
}

type branchesNode struct {
	repo    *git.Repository
	entries []branchesNodeEntry
}

func (n *branchesNode) Children() (map[string]fstree.Node, *fserror.Error) {
	children := map[string]fstree.Node{}
	for _, entry := range n.entries {
		switch len(entry.nameParts) {
		case 0:
			return nil, fserror.Unexpected(errors.Errorf("unexpected branch name: %s", entry.branchRef.Name()))

		case 1:
			branchCommit, err := n.repo.CommitObject(entry.branchRef.Hash())
			if err == plumbing.ErrObjectNotFound {
				return nil, fserror.Unexpected(errors.Errorf("Branch %s points to invalid or non-commit ref %s", entry.branchRef.Name(), entry.branchRef.Hash()))
			} else if err != nil {
				return nil, fserror.Unexpected(errors.Wrap(err, "find branch commit failed"))
			}

			branchTree, err := branchCommit.Tree()
			if err != nil {
				return nil, fserror.Unexpected(errors.Wrap(err, "find branch tree failed"))
			}

			children[entry.nameParts[0]] = &treeNode{repo: n.repo, tree: branchTree}

		default:
			var child *branchesNode
			if existingNode, ok := children[entry.nameParts[0]]; ok {
				if existingBranchesNode, ok := existingNode.(*branchesNode); ok {
					child = existingBranchesNode
				} else {
					return nil, fserror.Unexpected(errors.Errorf("conflicting parent/child branch name: %v", entry.branchRef.Name()))
				}
			} else {
				child = &branchesNode{repo: n.repo}
				children[entry.nameParts[0]] = child
			}

			child.entries = append(child.entries, branchesNodeEntry{nameParts: entry.nameParts[1:], branchRef: entry.branchRef})
		}
	}
	return children, nil
}

type treeNode struct {
	repo *git.Repository
	tree *object.Tree
}

func (n *treeNode) Children() (map[string]fstree.Node, *fserror.Error) {
	children := map[string]fstree.Node{}
	for i := range n.tree.Entries {
		treeEntry := &n.tree.Entries[i]
		switch treeEntry.Mode {
		case filemode.Dir:
			childTree, err := n.repo.TreeObject(treeEntry.Hash)
			if err != nil {
				return nil, fserror.Unexpected(err)
			}
			children[treeEntry.Name] = &treeNode{repo: n.repo, tree: childTree}

		case filemode.Regular, filemode.Executable:
			childFile, err := n.tree.TreeEntryFile(treeEntry)
			if err != nil {
				return nil, fserror.Unexpected(err)
			}
			children[treeEntry.Name] = &fileNode{file: childFile}

		default:
			// TODO(josh-newman): Use a logger from gitviewfs.
			fmt.Fprintf(os.Stderr, "skipping file mode %v: %s\n", treeEntry.Mode, treeEntry.Hash)
		}
	}
	return children, nil
}

type fileNode struct {
	file *object.File
}

func (n *fileNode) File() *object.File {
	return n.file
}
