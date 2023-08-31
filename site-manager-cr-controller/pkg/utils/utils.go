package utils

import (
	"fmt"
	"os"
)

// CheckFile is used to check, if file is defined and exists
func CheckFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file is undefined")
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("can't get info for %s: %s", filePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is not a file", filePath)
	}
	return nil
}

// Contains is supported is used to find if value v exists in array e
func Contains[E comparable](s []E, v E) bool {
	for _, e := range s {
		if v == e {
			return true
		}
	}
	return false
}
