package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func MD5File(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	md5 := md5.New()
	if _, err := io.Copy(md5, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(nil)), nil
}

func LocateFile(name string, subDirs []string) (path string, found bool) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// exPath is the directory this executable resides in

	exPath := filepath.Dir(ex)
	for _, subDir := range subDirs {
		searchPath := filepath.Join(exPath, subDir, name)
		matches, err := filepath.Glob(searchPath)
		if err != nil {
			panic(err)
		}
		if matches != nil {
			return matches[0], true
		}
	}
	return "", false
}
