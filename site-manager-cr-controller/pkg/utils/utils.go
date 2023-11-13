package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
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

// ParseYamlFile is used to parse given file to specified object
func ParseYamlFile(filePath string, obj interface{}) error {
	if err := CheckFile(filePath); err != nil {
		return err
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("can't read file %s: %s", filePath, err)
	}

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	err = decoder.Decode(obj)
	if err != nil {
		return fmt.Errorf("can't parse file content %s: %s", filePath, err)
	}
	return nil
}

// Contains is used to find if value v exists in array e
func Contains[E comparable](s []E, v E) bool {
	for _, e := range s {
		if v == e {
			return true
		}
	}
	return false
}

// FindFirstFromMap finds the first element in map which value meets specified filter func and returns its key pointer
func FindFirstFromMap[T comparable, V interface{}](objects map[T]V, filter func(V) bool) *T {
	for key, value := range objects {
		if filter(value) {
			return &key
		}
	}
	return nil
}

// FindFirstFromSlice finds the first element in slice which value meets specified filter func and returns its pointer
func FindFirstFromSlice[V interface{}](objects []V, filter func(V) bool) *V {
	for _, value := range objects {
		if filter(value) {
			return &value
		}
	}
	return nil
}
