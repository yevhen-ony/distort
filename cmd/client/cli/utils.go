package main

import (
	"errors"
	"os"
)

func EnsureFileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

func EnsureFileNotExists(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return os.ErrExist
}

func EnsureFolder(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return nil
	}
	return errors.New("not a dir")
}

func EnsureFolderOrNotExist(path string) (err error) {
	if err = EnsureFolder(path); err == nil {
		return nil
	}
	if err = EnsureFileNotExists(path); err == nil {
		return nil
	}
	return err
}
