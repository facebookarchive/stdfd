// Package stdfd provides routines for redirecting stdout and stderr.
package stdfd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func dup2file(fd int, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0700)); err != nil {
		return err
	}

	newfd, err := syscall.Open(
		path,
		syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT,
		0600,
	)
	if err != nil {
		return fmt.Errorf("could not open %s", path)
	}
	defer syscall.Close(newfd) // in case we fail the Dup2 below

	if err := syscall.Dup2(newfd, fd); err != nil {
		return err
	}

	if err := syscall.Close(newfd); err != nil {
		return err
	}

	return nil
}

// RedirectOutputs (os.Stdout & os.Stderr) if necessary. There are some
// special implications here:
//
// - Both arguments are paths.
//
// - Blank arguments are ignored (that is, no redirection is performed).
//
// - If paths are the same they are redirected to the same file descriptor.
//
// - Parent directories are created if necessary.
//
// - Target file is created if necessary (will be appended to if exists).
func RedirectOutputs(stdout, stderr string) error {
	// no redirection
	if stderr == "" && stdout == "" {
		return nil
	}

	// redirecting stdout
	if stdout != "" {
		if err := dup2file(1, stdout); err != nil {
			return err
		}
	}

	// redirecting stderr to the same file
	if stderr == stdout {
		if err := syscall.Dup2(1, 2); err != nil {
			return err
		}
		return nil
	}

	// redirecting stderr
	if stderr != "" {
		if err := dup2file(2, stderr); err != nil {
			return err
		}
	}

	return nil
}
