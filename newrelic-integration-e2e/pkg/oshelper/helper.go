package oshelper

import (
	"fmt"
	"io"
	"os"
)

func RemoveDirectories(dirs ...string) error {
	for i := range dirs {
		dir := dirs[i]
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destination.Close()
	}()
	_, err = io.Copy(destination, source)
	return err
}

func MakeDirs(perm os.FileMode, dirs ...string) error {
	for i := range dirs {
		dir := dirs[i]
		if err := os.Mkdir(dir, perm); err != nil {
			return err
		}
	}
	return nil
}
