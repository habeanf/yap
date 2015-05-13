package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
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
