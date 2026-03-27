package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BlobStore abstracts object persistence (disk today, S3-compatible later).
// Keys are logical paths (e.g. surveys/{surveyID}/{mediaID}.bin); same key can be an S3 object name.
type BlobStore interface {
	Put(ctx context.Context, key string, r io.Reader, maxBytes int64) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Remove(ctx context.Context, key string) error
}

// DiskStore writes under rootDir.
type DiskStore struct {
	rootDir string
}

// NewDiskStore returns a store rooted at rootDir (must be absolute or cwd-relative).
func NewDiskStore(rootDir string) (*DiskStore, error) {
	if rootDir == "" {
		return nil, fmt.Errorf("storage: empty root dir")
	}
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}
	return &DiskStore{rootDir: abs}, nil
}

func sanitizeKey(key string) error {
	if key == "" || strings.Contains(key, "..") || filepath.IsAbs(key) || strings.HasPrefix(key, string(os.PathSeparator)) {
		return fmt.Errorf("storage: invalid key")
	}
	clean := filepath.Clean(key)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return fmt.Errorf("storage: invalid key")
	}
	return nil
}

func (d *DiskStore) fullPath(key string) (string, error) {
	if err := sanitizeKey(key); err != nil {
		return "", err
	}
	return filepath.Join(d.rootDir, filepath.FromSlash(key)), nil
}

// Put writes the stream to root/key, creating parent directories. At most maxBytes are read.
func (d *DiskStore) Put(ctx context.Context, key string, r io.Reader, maxBytes int64) error {
	if maxBytes <= 0 {
		return fmt.Errorf("storage: maxBytes must be positive")
	}
	path, err := d.fullPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	lr := io.LimitReader(r, maxBytes+1)
	n, copyErr := io.Copy(f, lr)
	_ = f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if n > maxBytes {
		_ = os.Remove(tmp)
		return fmt.Errorf("storage: payload exceeds max size")
	}
	select {
	case <-ctx.Done():
		_ = os.Remove(tmp)
		return ctx.Err()
	default:
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Open returns a read closer for the object at key.
func (d *DiskStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	_ = ctx
	path, err := d.fullPath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return f, nil
}

// Remove deletes the object at key; missing file is not an error.
func (d *DiskStore) Remove(ctx context.Context, key string) error {
	_ = ctx
	path, err := d.fullPath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
