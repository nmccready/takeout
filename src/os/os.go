package os

import (
	"io"
	"os"

	"github.com/nmccready/takeout/src/logger"
)

var debug = logger.Spawn("os")

// Go OS has Rename, but not Copy, so we make it happen
func Copy(src, dest string) error {
	debug.Log("Copy: %s, %s", src, dest)
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = srcFile.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}
