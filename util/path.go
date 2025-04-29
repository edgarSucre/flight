package util

import (
	"os"
	"path"
)

func FilePath(fileName string) string {
	fullPath, _ := os.Getwd()

	// for local debugging
	if os.Getenv("debug") == "true" {
		fullPath = path.Join(fullPath, "..")
	}

	return path.Join(fullPath, fileName)
}
