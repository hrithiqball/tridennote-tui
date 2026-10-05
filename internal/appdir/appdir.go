package appdir

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	name       = "tridennote"
	legacyName = "cerebrum"
)

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, name)
	legacy := filepath.Join(base, legacyName)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if info, err := os.Stat(legacy); err == nil && info.IsDir() {
			_ = os.Rename(legacy, dir)
		}
	}
	return dir, nil
}

func File(file string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, file), nil
}
