package platforms

import (
	"os"
)

// createFile is a helper to create a file for writing
func createFile(path string) (*os.File, error) {
	return os.Create(path)
}
