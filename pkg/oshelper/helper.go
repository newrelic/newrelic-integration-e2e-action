package oshelper

import (
	"fmt"
	"io"
	"os"
)

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
	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	if err := os.Chmod(destination.Name(), sourceFileStat.Mode()); err != nil {
		return err
	}

	return nil
}
