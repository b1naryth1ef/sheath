package fsx

import (
	"io/fs"
	"os"
	"time"
)

// FileInfoSpec contains the metadata required to implement `fs.FileInfo`
type FileInfoSpec struct {
	Name    string
	Size    int64
	Mode    fs.FileMode
	IsDir   bool
	Sys     any
	ModTime time.Time
}

// FileInfo implements a basic version of `fs.FileInfo`
type FileInfo struct {
	Spec FileInfoSpec
}

// NewFileInfo creates a new `FileInfo` from a name and size
func NewFileInfo(name string, size int) *FileInfo {
	return &FileInfo{
		Spec: FileInfoSpec{
			Name:    name,
			Size:    int64(size),
			Mode:    os.ModePerm,
			ModTime: time.Now(),
		},
	}
}

func (f *FileInfo) Name() string {
	return f.Spec.Name
}

func (f *FileInfo) Size() int64 {
	return f.Spec.Size
}

func (f *FileInfo) Mode() fs.FileMode {
	return f.Spec.Mode
}

func (f *FileInfo) IsDir() bool {
	return f.Spec.IsDir
}

func (f *FileInfo) Sys() any {
	return f.Spec.Sys
}

func (f *FileInfo) ModTime() time.Time {
	return f.Spec.ModTime
}
