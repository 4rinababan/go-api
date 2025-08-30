package utils

import (
	"os"
	"strconv"
)

func StringToInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
func EnsureDir(dirName string) (string, error) {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return "", err
	}
	return dirName, nil
}
