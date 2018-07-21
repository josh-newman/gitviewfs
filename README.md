# gitviewfs

FUSE filesystem that provides a read-only view into a git repository's branches and their contents.

It was written for fun and should be considered experimental.

## Installation

First [install Go](https://golang.org/doc/install), then gitviewfs:
```bash
$ go get -u github.com/josh-newman/gitviewfs/cmd/gitviewfs/
```
You'll also need to install FUSE ([macOS](https://osxfuse.github.io/), [Linux](https://github.com/libfuse/libfuse)).

## Usage

gitviewfs has two required arguments:
```bash
$ gitviewfs [-debug] /mount/point /path/to/repository
```
For example:
```bash
$ mkdir /tmp/view

$ gitviewfs /tmp/view $GOPATH/src/github.com/josh-newman/gitviewfs &

$ ls /tmp/view/refs/heads/master/
Gopkg.lock      Gopkg.toml      README.md       cmd             gitviewfs

$ head -n 1 /tmp/view/refs/heads/master/README.md
# gitviewfs
```

## TODO

* Figure out if pathfs function implementations should pay attention to `fuse.Context`. Should it
  implement some access control?
* Consider using [nodefs](https://github.com/hanwen/go-fuse/tree/master/fuse/nodefs) instead of
  pathfs; gitviewfs already has a node abstraction internally.
* Do something with git submodules (show their contents recursively?).
* Access times.
* Memory-efficient file reading.
* Support Git LFS.
* Consider adding FUSE options for mounting only some branches, etc.
* Consider adding FUSE view into git history, in some way.
* Tests.
