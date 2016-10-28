package toolbox

import (
	"fmt"
	"os"
)

var dirMode os.FileMode = 0644

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

// CreateDirIfNotExist creates directory if they do not exist
func CreateDirIfNotExist(dirs ...string) error {
	for _, dir := range dirs {
		if !FileExists(dir) {
			err := os.Mkdir(dir, dirMode)
			if err != nil {
				return fmt.Errorf("Failed to create dir %v %v", dir, err)
			}
		}
	}
	return nil
}
