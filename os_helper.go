package toolbox

import (
	"fmt"
	"os"
	"path"
	"strings"
)

var dirMode os.FileMode = 0744

// RemoveFileIfExist remove file if exists
func RemoveFileIfExist(filenames ...string) error {
	for _, filename := range filenames {
		if !FileExists(filename) {
			continue
		}
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

// FileExists checks if file exists
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}

// IsDirectory checks if file is directory
func IsDirectory(location string) bool {
	if stat, _ := os.Stat(location); stat != nil {
		return stat.IsDir()
	}
	return false
}

// CreateDirIfNotExist creates directory if they do not exist
func CreateDirIfNotExist(dirs ...string) error {
	for _, dir := range dirs {
		if len(dir) > 1 && strings.HasSuffix(dir, "/") {
			dir = dir[:len(dir)-1]
		}
		parent, _ := path.Split(dir)
		if parent != "/" && parent != dir {
			CreateDirIfNotExist(parent)
		}
		if !FileExists(dir) {
			err := os.Mkdir(dir, dirMode)
			if err != nil {
				return fmt.Errorf("failed to create dir %v %v", dir, err)
			}
		}
	}
	return nil
}
