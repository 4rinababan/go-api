package utils

import (
	"os"
)

// DeleteFile tries to delete a file from the filesystem
func DeleteFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.Remove(path)
	}
	return nil // file tidak ada, aman
}
