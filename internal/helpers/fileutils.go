package helpers

import (
	"fmt"
	"io"
	"os"
)

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// https://github.com/probonopd/go-appimage/blob/23ad67c727fb762867fe96db06d600a7cdaf297d/internal/helpers/helpers.go#L150

// CheckIfFileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
// Returns true if it does, false otherwise.
func CheckIfFileExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) || info.IsDir() {
		return false
	}
	// anything else is not good
	// https://stackoverflow.com/q/12518876
	// Schrodinger: file may or may not exist. See err for details.
	return true

}

// CheckIfSymlinkExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
// Returns true if it does, false otherwise.
func CheckIfSymlinkExists(filepath string) bool {
	info, err := os.Lstat(filepath)
	if info == nil {
		return true
	}
	if info.IsDir() {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// anything else is not good
	// https://stackoverflow.com/q/12518876
	// Schrodinger: file may or may not exist. See err for details.
	return true

}

// CheckIfDirectoryExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
// Returns true if it does, false otherwise.
func CheckIfDirectoryExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
