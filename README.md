# gitviewfs

Work-in-progress FUSE filesystem that provides a read-only view into a git repository's branches.

## TODO

* Figure out if pathfs function implementations should pay attention to `fuse.Context`. Should it
  implement some access control?
* Consider using [nodefs](https://github.com/hanwen/go-fuse/tree/master/fuse/nodefs) instead of
  pathfs; gitviewfs already has a node abstraction internally.
* Handle symlinks nicely.
* Do something with git submodules.
* Access times.
