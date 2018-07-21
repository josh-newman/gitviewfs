package gitfstree

import (
	"fmt"
	"github.com/josh-newman/gitviewfs/gitviewfs/fserror"
	"github.com/josh-newman/gitviewfs/gitviewfs/fstree"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"os"
	"strings"
)

func New(repo *git.Repository) (fstree.Node, error) {
	refs, err := repo.References()
	defer refs.Close()
	if err != nil {
		return nil, errors.Wrap(err, "list references failed")
	}

	node := referencesNode{repo: repo}
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		nameParts := strings.Split(string(ref.Name()), "/")
		node.entries = append(node.entries, referencesNodeEntry{nameParts: nameParts, ref: ref})
		return nil
	})
	if err != nil {
		return nil, fserror.Unexpected(errors.Wrap(err, "processing references failed"))
	}

	return &node, nil
}

type referencesNodeEntry struct {
	nameParts []string
	ref       *plumbing.Reference
}

type referencesNode struct {
	repo    *git.Repository
	entries []referencesNodeEntry
}

func (n *referencesNode) Children() (map[string]fstree.Node, *fserror.Error) {
	children := map[string]fstree.Node{}
	for _, entry := range n.entries {
		switch len(entry.nameParts) {
		case 0:
			return nil, fserror.Unexpected(errors.Errorf("unexpected ref name: %s", entry.ref.Name()))

		case 1:
			hash := entry.ref.Hash()
			if hash == plumbing.ZeroHash {
				// HEAD references (in refs/heads/* and refs/remotes/.../*) seem to be listed with hash
				// zero, so we skip them.
				continue
			}
			refCommit, err := n.repo.CommitObject(hash)
			if err == plumbing.ErrObjectNotFound {
				return nil, fserror.Unexpected(errors.Errorf("Ref name %s points to invalid or non-commit ref %s", entry.ref.Name(), entry.ref.Hash()))
			} else if err != nil {
				return nil, fserror.Unexpected(errors.Wrap(err, "find ref commit failed"))
			}

			refTree, err := refCommit.Tree()
			if err != nil {
				return nil, fserror.Unexpected(errors.Wrap(err, "find ref tree failed"))
			}

			children[entry.nameParts[0]] = &treeNode{repo: n.repo, tree: refTree}

		default:
			var child *referencesNode
			if existingNode, ok := children[entry.nameParts[0]]; ok {
				if existingBranchesNode, ok := existingNode.(*referencesNode); ok {
					child = existingBranchesNode
				} else {
					return nil, fserror.Unexpected(errors.Errorf("conflicting parent/child branch name: %v", entry.ref.Name()))
				}
			} else {
				child = &referencesNode{repo: n.repo}
				children[entry.nameParts[0]] = child
			}

			child.entries = append(child.entries, referencesNodeEntry{nameParts: entry.nameParts[1:], ref: entry.ref})
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

		case filemode.Regular, filemode.Executable, filemode.Symlink:
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
