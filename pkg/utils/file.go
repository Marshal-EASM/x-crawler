package utils

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return !info.IsDir()
}
func ReadFile(filename string) (content []string, err error) {
	if !FileExists(filename) {
		return content, errors.New("file doesn't exist")
	}
	fileData, err := ioutil.ReadFile(filename)
	if err == nil {
		return strings.Split(string(fileData), "\n"), nil
	} else {
		return content, err
	}
}
