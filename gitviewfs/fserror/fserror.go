package fserror

import (
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
)

// Error represents either a normal or an unexpected FUSE error.
type Error struct {
	Status fuse.Status
	// UnexpectedErr should be nil for normal errors (like file not found, etc.).
	UnexpectedErr error
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	return fmt.Sprintf("fuse error: %s: %s", e.Status, e.UnexpectedErr)
}

func Expected(status fuse.Status) *Error {
	return &Error{status, nil}
}

func Unexpected(err error) *Error {
	return &Error{fuse.EIO, err}
}
