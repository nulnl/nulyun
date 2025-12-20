package files

import (
	"context"
	"os"
	"path"

	"github.com/spf13/afero"
)

type noopCache struct{}

func (n *noopCache) Store(ctx context.Context, key string, value []byte) error {
	return nil
}

func (n *noopCache) Load(ctx context.Context, key string) (value []byte, exist bool, err error) {
	return nil, false, nil
}

func (n *noopCache) Delete(ctx context.Context, key string) error {
	return nil
}

// NewNoOp returns a no-op cache implementation.
func NewNoOp() Interface {
	return &noopCache{}
}

type fileCache struct {
	fs  afero.Fs
	dir string
}

func (f *fileCache) Store(ctx context.Context, key string, value []byte) error {
	if err := f.fs.MkdirAll(f.dir, 0700); err != nil {
		return err
	}
	p := path.Join(f.dir, key)
	return afero.WriteFile(f.fs, p, value, 0600)
}

func (f *fileCache) Load(ctx context.Context, key string) (value []byte, exist bool, err error) {
	p := path.Join(f.dir, key)
	b, err := afero.ReadFile(f.fs, p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return b, true, nil
}

func (f *fileCache) Delete(ctx context.Context, key string) error {
	p := path.Join(f.dir, key)
	return f.fs.Remove(p)
}

// New creates a simple file-backed cache in the given directory.
func New(afs afero.Fs, dir string) Interface {
	return &fileCache{fs: afs, dir: dir}
}
