# gitviewfs

Work-in-progress FUSE filesystem that provides a read-only view into a git repository's branches.

## TODO

* Documentation on how to use gitviewfs.
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
