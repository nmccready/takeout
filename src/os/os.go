package os

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/nmccready/takeout/src/internal/logger"
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

func ExitOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("env var %s is not set", key)
	}
	return val, nil
}

func GetRequiredEnv(key string) string {
	val, err := GetEnv(key)
	if err != nil {
		panic(err)
	}
	return val
}

func LoadJSON(paths []string, v interface{}) error {
	bytes, err := os.ReadFile(path.Join(paths...))

	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, v)
}
