// automatically generated by stateify.

package nsfs

import (
	"gvisor.dev/gvisor/pkg/state"
)

func (r *inodeRefs) StateTypeName() string {
	return "pkg/sentry/fsimpl/nsfs.inodeRefs"
}

func (r *inodeRefs) StateFields() []string {
	return []string{
		"refCount",
	}
}

func (r *inodeRefs) beforeSave() {}

// +checklocksignore
func (r *inodeRefs) StateSave(stateSinkObject state.Sink) {
	r.beforeSave()
	stateSinkObject.Save(0, &r.refCount)
}

// +checklocksignore
func (r *inodeRefs) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &r.refCount)
	stateSourceObject.AfterLoad(r.afterLoad)
}

func (f *filesystemType) StateTypeName() string {
	return "pkg/sentry/fsimpl/nsfs.filesystemType"
}

func (f *filesystemType) StateFields() []string {
	return []string{}
}

func (f *filesystemType) beforeSave() {}

// +checklocksignore
func (f *filesystemType) StateSave(stateSinkObject state.Sink) {
	f.beforeSave()
}

func (f *filesystemType) afterLoad() {}

// +checklocksignore
func (f *filesystemType) StateLoad(stateSourceObject state.Source) {
}

func (fs *filesystem) StateTypeName() string {
	return "pkg/sentry/fsimpl/nsfs.filesystem"
}

func (fs *filesystem) StateFields() []string {
	return []string{
		"Filesystem",
		"devMinor",
	}
}

func (fs *filesystem) beforeSave() {}

// +checklocksignore
func (fs *filesystem) StateSave(stateSinkObject state.Sink) {
	fs.beforeSave()
	stateSinkObject.Save(0, &fs.Filesystem)
	stateSinkObject.Save(1, &fs.devMinor)
}

func (fs *filesystem) afterLoad() {}

// +checklocksignore
func (fs *filesystem) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &fs.Filesystem)
	stateSourceObject.Load(1, &fs.devMinor)
}

func (i *Inode) StateTypeName() string {
	return "pkg/sentry/fsimpl/nsfs.Inode"
}

func (i *Inode) StateFields() []string {
	return []string{
		"InodeAttrs",
		"InodeAnonymous",
		"InodeNotDirectory",
		"InodeNotSymlink",
		"InodeWatches",
		"inodeRefs",
		"locks",
		"namespace",
		"mnt",
	}
}

func (i *Inode) beforeSave() {}

// +checklocksignore
func (i *Inode) StateSave(stateSinkObject state.Sink) {
	i.beforeSave()
	stateSinkObject.Save(0, &i.InodeAttrs)
	stateSinkObject.Save(1, &i.InodeAnonymous)
	stateSinkObject.Save(2, &i.InodeNotDirectory)
	stateSinkObject.Save(3, &i.InodeNotSymlink)
	stateSinkObject.Save(4, &i.InodeWatches)
	stateSinkObject.Save(5, &i.inodeRefs)
	stateSinkObject.Save(6, &i.locks)
	stateSinkObject.Save(7, &i.namespace)
	stateSinkObject.Save(8, &i.mnt)
}

func (i *Inode) afterLoad() {}

// +checklocksignore
func (i *Inode) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &i.InodeAttrs)
	stateSourceObject.Load(1, &i.InodeAnonymous)
	stateSourceObject.Load(2, &i.InodeNotDirectory)
	stateSourceObject.Load(3, &i.InodeNotSymlink)
	stateSourceObject.Load(4, &i.InodeWatches)
	stateSourceObject.Load(5, &i.inodeRefs)
	stateSourceObject.Load(6, &i.locks)
	stateSourceObject.Load(7, &i.namespace)
	stateSourceObject.Load(8, &i.mnt)
}

func (fd *namespaceFD) StateTypeName() string {
	return "pkg/sentry/fsimpl/nsfs.namespaceFD"
}

func (fd *namespaceFD) StateFields() []string {
	return []string{
		"FileDescriptionDefaultImpl",
		"LockFD",
		"vfsfd",
		"inode",
	}
}

func (fd *namespaceFD) beforeSave() {}

// +checklocksignore
func (fd *namespaceFD) StateSave(stateSinkObject state.Sink) {
	fd.beforeSave()
	stateSinkObject.Save(0, &fd.FileDescriptionDefaultImpl)
	stateSinkObject.Save(1, &fd.LockFD)
	stateSinkObject.Save(2, &fd.vfsfd)
	stateSinkObject.Save(3, &fd.inode)
}

func (fd *namespaceFD) afterLoad() {}

// +checklocksignore
func (fd *namespaceFD) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &fd.FileDescriptionDefaultImpl)
	stateSourceObject.Load(1, &fd.LockFD)
	stateSourceObject.Load(2, &fd.vfsfd)
	stateSourceObject.Load(3, &fd.inode)
}

func init() {
	state.Register((*inodeRefs)(nil))
	state.Register((*filesystemType)(nil))
	state.Register((*filesystem)(nil))
	state.Register((*Inode)(nil))
	state.Register((*namespaceFD)(nil))
}
