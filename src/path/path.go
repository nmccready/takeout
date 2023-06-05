package path

import (
	"errors"
	"path/filepath"
)

// Runtime Caller return _ uintptr, filename string, _ int, ok bool

// __filename equivalent to nodejs
func Filename(_ uintptr, filename string, _ int, ok bool) (string, error) {
	if !ok {
		return "", errors.New("unable to get filename from caller")
	}
	return filename, nil
}

func FilenameForce(pc uintptr, filename string, line int, ok bool) string {
	//nolint
	ret, _ := Filename(pc, filename, line, ok)
	return ret
}

// __dirname equivalent to nodejs
func Dirname(filename string) (string, error) {
	return filepath.Dir(filename), nil
}

func DirnameForce(filename string) string {
	//nolint
	ret, _ := Dirname(filename)
	return ret
}
