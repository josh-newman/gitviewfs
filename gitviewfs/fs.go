package gitviewfs

import (
	"gopkg.in/src-d/go-git.v4"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/fuse"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"log"
	"github.com/pkg/errors"
	"io"
	"strings"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"io/ioutil"
	"os"
)

type gitviewfs struct {
	pathfs.FileSystem
	repo *git.Repository
	logger *log.Logger
}

func New(repo *git.Repository) pathfs.FileSystem {
	return &gitviewfs{
		FileSystem: pathfs.NewDefaultFileSystem(),
		repo: repo,
		logger: log.New(ioutil.Discard, "gitviewfs", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC),
	}
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
	tree, entry, ferr := f.findTreeEntry(name)
	if ferr != nil {
		log.Print(ferr.unexpectedErr)
		return nil, ferr.status
	}

	var attr fuse.Attr
	switch entry.Mode {
	case filemode.Dir:
		attr.Mode = fuse.S_IFDIR | 0555
	case filemode.Regular:
		attr.Mode = fuse.S_IFREG | 0444
		file, err := tree.TreeEntryFile(entry)
		if err != nil {
			log.Print(err)
			return nil, fuse.EIO
		}
		attr.Size = uint64(file.Size)
	default:
		log.Printf("skipping file with mode: %s", entry.Mode)
		return nil, fuse.ENOENT
	}

	return &attr, fuse.OK
}

// chooseBranchTree inspects a path and returns a tree entry from the corresponding branch
// corresponding to path. Returns fuse.OK on success or an error code on failure.
func (f *gitviewfs) findTreeEntry(path string) (*object.Tree, *object.TreeEntry, *fsError) {
	branchRefs, err := f.repo.Branches()
	if err != nil {
		return nil, nil, newUnexpectedFsError(errors.Wrap(err, "list branches failed"))
	}

	for {
		branchRef, err := branchRefs.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, newUnexpectedFsError(errors.Wrap(err, "next branch failed"))
		}

		splitPath := strings.Split(path, "/")
		branchPath := strings.Split(string(branchRef.Name()), "/")
		if isPrefixSlice(branchPath, splitPath) {
			branchCommit, err := f.repo.CommitObject(branchRef.Hash())
			if err == plumbing.ErrObjectNotFound {
				return nil, nil, newUnexpectedFsError(errors.Errorf("Branch %s points to invalid or non-commit ref %s", branchRef.Name(), branchRef.Hash()))
			} else if err != nil {
				return nil, nil, newUnexpectedFsError(errors.Wrap(err, "find branch commit failed"))
			}

			branchTree, err := branchCommit.Tree()
			if err != nil {
				return nil, nil, newUnexpectedFsError(errors.Wrap(err, "find branch tree failed"))
			}

			remainingPath := splitPath[len(branchPath):]
			entry, err := branchTree.FindEntry(strings.Join(remainingPath, "/"))
			if err == object.ErrDirectoryNotFound {
				return nil, nil, newNormalFsError(fuse.ENOENT)
			} else if err != nil {
				return nil, nil, newUnexpectedFsError(errors.Wrap(err, "choosePathTree failed"))
			}

			return branchTree, entry, nil
		}
	}

	// No matching branch found.
	return nil, nil, newNormalFsError(fuse.ENOENT)
}

// isPrefixSlice returns true if maybePrefix is a prefix of other, according to comparing
// corresponding slice elements.
func isPrefixSlice(maybePrefix, other []string) bool {
	if len(maybePrefix) > len(other) {
		return false
	}
	for i, p := range maybePrefix {
		if p != other[i] {
			return false
		}
	}
	return true
}
