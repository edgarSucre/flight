package util

import (
	"os"
	"path"
)

func FilePath(fileName string) string {
	fullPath, _ := os.Getwd()

	// TODO: remove this
	if os.Getenv("debug") == "true" {
		fullPath = path.Join(fullPath, "..")
	}

	return path.Join(fullPath, fileName)
}
