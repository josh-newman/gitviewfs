package gitviewfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"fmt"
)

// fsError represents either a normal or an unexpected FUSE error.
type fsError struct {
	status fuse.Status
	// unexpectedErr should be nil for normal errors (like file not found, etc.).
	unexpectedErr error
}

var _ error = (*fsError)(nil)

func (e *fsError) Error() string {
	return fmt.Sprintf("fuse error: %s: %s", e.status, e.unexpectedErr)
}

func newNormalFsError(status fuse.Status) *fsError {
	return &fsError{status, nil}
}

func newUnexpectedFsError(err error) *fsError {
	return &fsError{fuse.EIO, err}
}
