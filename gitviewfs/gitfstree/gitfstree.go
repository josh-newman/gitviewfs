package gitfstree

import (
	"gopkg.in/src-d/go-git.v4"
	"github.com/pkg/errors"
	"io"
	"github.com/josh-newman/git-view-fs/gitviewfs/fserror"
	"github.com/josh-newman/git-view-fs/gitviewfs/fstree"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"strings"
	"gopkg.in/src-d/go-git.v4/plumbing"
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
	repo *git.Repository
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

			children[entry.nameParts[0]] = &treeNode{tree: branchTree}

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

// TODO(josh-newman): Real implementation.
type treeNode struct {
	tree *object.Tree
}

func (n *treeNode) Size() uint64 {
	return 0
}

func (n *treeNode) Executable() bool {
	return false
}
