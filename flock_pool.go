// +build linux

package lockd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"syscall"
	"time"
)

// TODO(teawithsand): make this work on other platforms and test it

// FilePoolMutex creates pool of Size files, which are empty.
//
// It uses call to syscall.Flock in order to lock files.
// This way lock is automatically freed after process crashes, which is not guaranteed by FileMutex.
type FilePoolMutex struct {
	Dir  string
	Size int

	Hasher Hasher
}

type filePoolLocker struct {
	Path string
	err  error
	file *os.File
}

func (fpl *filePoolLocker) Lock(ctx context.Context) (err error) {
	if fpl.err != nil {
		return fpl.err
	}

	// TODO(teawithsand): better implementation with something like
	// https://github.com/karelzak/util-linux/blob/master/sys-utils/flock.c
	for {
		if fpl.file == nil {
			f, ferr := os.OpenFile(fpl.Path, os.O_CREATE|os.O_RDWR, 0666)
			if ferr != nil {
				err = ferr
				return
			}

			fpl.file = f
		}

		err = syscall.Flock(int(fpl.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if errors.Is(err, syscall.EWOULDBLOCK) {
			t := time.NewTimer(time.Millisecond * 100) // TODO(teawtihsand): allow tweaking this? Or some algorithm, which increases delay after each try?
			select {
			case <-ctx.Done():
				t.Stop()
				err = ctx.Err()
				return
			case <-t.C:
				t.Stop()
			}

			continue
		} else if errors.Is(err, syscall.EINTR) {
			continue
		} else if err != nil {
			return
		}

		break
	}

	return
}

func (fpl *filePoolLocker) Unlock(ctx context.Context) (err error) {
	if fpl.err != nil {
		return fpl.err
	}

	if fpl.file != nil {
		// ignore errors? Should work as is, because it's tear-down procedure
		// it calls close...
		for {
			ferr := syscall.Flock(int(fpl.file.Fd()), syscall.LOCK_UN)
			if errors.Is(ferr, syscall.EINTR) {
				continue
			}

			break
		}
		fpl.file.Close()
		fpl.file = nil
	}
	return
}

// GetLock gets lock for specified key/id.
func (fm *FilePoolMutex) GetLock(key string) ContextLocker {
	var filename string
	if fm.Hasher != nil {
		filename = fmt.Sprintf("%d", fm.Hasher(key))
	} else {
		filename = fmt.Sprintf("%d", DefaultHasher(key))
	}
	return &fileLocker{
		Path: path.Join(fm.Dir, filename),
	}
}
