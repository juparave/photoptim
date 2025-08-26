package remotefs

import (
	"context"
	"io"
	"io/fs"
	"time"
)

// ConnectionConfig holds remote connection parameters.
type ConnectionConfig struct {
	Host       string
	Port       int
	User       string
	Password   string // optional fallback
	KeyPath    string // optional private key path
	RemotePath string // initial chroot path
}

// RemoteEntry describes a file or directory on the remote system.
type RemoteEntry struct {
	Path       string
	Name       string
	Size       int64
	Mode       fs.FileMode
	ModTime    time.Time
	IsDir      bool
	Symlink    bool
	TargetPath string // resolved target if symlink (within chroot)
}

// RemoteFS is a protocol‑agnostic filesystem abstraction.
type RemoteFS interface {
	Connect(ctx context.Context, cfg ConnectionConfig) error
	Close() error
	List(ctx context.Context, path string) ([]RemoteEntry, error)
	Stat(ctx context.Context, path string) (RemoteEntry, error)
	Open(ctx context.Context, path string) (io.ReadCloser, RemoteEntry, error)
	Create(ctx context.Context, path string, overwrite bool) (io.WriteCloser, error)
	Join(elem ...string) string
	Root() string
}
