package toolbox

import (
	"fmt"
	"os"
)

var dirMode os.FileMode = 0644

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

func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}

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
