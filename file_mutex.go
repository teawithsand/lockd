package lockd

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"
)

// FileMutex uses os.OpenFile with os.O_EXCL flag.
// It uses hashing in order to allow arbitrary names of keys.
// User can provide NameSanitizer in order to implement own logic of translating ids to filenames without path traversal.
//
// It creates file on lock, and removes it on unlock.
// It's safe to use it across multiple processes, unless different processes have different hashing parameters/name sanitizer.
//
// Note 1: this lock is not safe against process crashes. If process crashes then lock stays acquired forever.
// Note 2: This lock should work on NFSv3, however due to it's crash-unsafety it's probably poor choice. http://nfs.sourceforge.net/#faq_d10
// Note 3: This lock may also result in invalid state in case of IO failure on non-tmpfs drives.
type FileMutex struct {
	Dir string

	Hasher        Hasher
	NameSanitizer func(id string) (filename string, err error)
}

type fileLocker struct {
	Path string
	err  error
}

func (f *fileLocker) Lock(ctx context.Context) (err error) {
	if f.err != nil {
		return f.err
	}

	for {
		f, ferr := os.OpenFile(f.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
		if ferr == os.ErrExist {
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
		} else if ferr != nil {
			err = ferr
			return
		}
		f.Close()

		break
	}

	return
}

func (f *fileLocker) Unlock(ctx context.Context) (err error) {
	if f.err != nil {
		return f.err
	}
	err = os.Remove(f.Path)
	if err == os.ErrNotExist { // ignore err not exist?
		err = nil
	}
	return
}

// GetLock gets lock for specified key/id.
func (fm *FileMutex) GetLock(key string) ContextLocker {
	var filename string
	var err error
	if fm.NameSanitizer != nil {
		filename, err = fm.NameSanitizer(key)
	} else if fm.Hasher != nil {
		filename = fmt.Sprintf("%d", fm.Hasher(key))
	} else {
		filename = fmt.Sprintf("%d", DefaultHasher(key))
	}
	return &fileLocker{
		Path: path.Join(fm.Dir, filename),
		err:  err,
	}
}
