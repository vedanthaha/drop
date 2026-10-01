package platforms

import (
	"os"
)

// createFile is a helper to create a file for writing
func createFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
}
