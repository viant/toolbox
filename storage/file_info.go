package storage

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (i *fileInfo) Name() string {
	return i.name
}

func (i *fileInfo) Size() int64 {
	return i.size
}
func (i *fileInfo) Mode() os.FileMode {
	return i.mode
}
func (i *fileInfo) ModTime() time.Time {
	return i.modTime
}

func (i *fileInfo) IsDir() bool {
	return i.isDir
}

func (i *fileInfo) Sys() interface{} {
	return i
}

func NewFileInfo(name string, size int64, mode os.FileMode, modificationTime time.Time, isDir bool) os.FileInfo {
	return &fileInfo{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modificationTime,
		isDir:   isDir,
	}
}

func NewFileMode(fileAttributes string) (os.FileMode, error) {
	var result os.FileMode
	if len(fileAttributes) != 10 {
		return result, fmt.Errorf("Invalid attribute length %v %v", fileAttributes, len(fileAttributes))
	}

	const fileType = "dalTLDpSugct"
	var fileModePosition = strings.Index(fileType, string(fileAttributes[0]))
	if fileModePosition != -1 {
		result = 1 << uint(32-1-fileModePosition)
	}

	const filePermission = "rwxrwxrwx"
	for i, c := range filePermission {
		if c == rune(fileAttributes[i+1]) {
			result = result | 1<<uint(9-1-i)
		}
	}
	return result, nil

}
