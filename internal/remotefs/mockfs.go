package remotefs

import (
    "bytes"
    "context"
    "io"
    "io/fs"
    "path/filepath"
    "sync"
    "time"
)

type MockFS struct {
    root string
    mu   sync.Mutex
    files map[string]*mockFile
}

type mockFile struct {
    data []byte
    mode fs.FileMode
    modTime time.Time
}

func NewMockFS(root string) *MockFS { return &MockFS{root:root, files: map[string]*mockFile{}} }
func (m *MockFS) Connect(ctx context.Context, cfg ConnectionConfig) error { return nil }
func (m *MockFS) Close() error { return nil }
func (m *MockFS) Root() string { return m.root }
func (m *MockFS) Join(elem ...string) string { return filepath.Join(elem...) }

func (m *MockFS) put(path string, data []byte) { m.mu.Lock(); defer m.mu.Unlock(); m.files[path] = &mockFile{data: append([]byte(nil), data...), mode:0o644, modTime: time.Now()} }

func (m *MockFS) List(ctx context.Context, path string) ([]RemoteEntry, error) {
    m.mu.Lock(); defer m.mu.Unlock()
    out := []RemoteEntry{}
    for p,f := range m.files { if filepath.Dir(p)==path { out = append(out, RemoteEntry{Path:p, Name:filepath.Base(p), Size:int64(len(f.data)), Mode:f.mode, ModTime:f.modTime}) } }
    return out, nil
}
func (m *MockFS) Stat(ctx context.Context, path string) (RemoteEntry, error) {
    m.mu.Lock(); defer m.mu.Unlock(); f:=m.files[path]; if f==nil { return RemoteEntry{}, fs.ErrNotExist }; return RemoteEntry{Path:path, Name:filepath.Base(path), Size:int64(len(f.data)), Mode:f.mode, ModTime:f.modTime}, nil }

type nopCloser struct { *bytes.Reader }
func (n nopCloser) Close() error { return nil }

func (m *MockFS) Open(ctx context.Context, path string) (io.ReadCloser, RemoteEntry, error) {
    m.mu.Lock(); f:=m.files[path]; m.mu.Unlock(); if f==nil { return nil, RemoteEntry{}, fs.ErrNotExist }
    r := nopCloser{bytes.NewReader(f.data)}
    e := RemoteEntry{Path:path, Name:filepath.Base(path), Size:int64(len(f.data)), Mode:f.mode, ModTime:f.modTime}
    return r, e, nil
}

type writeBuffer struct { bytes.Buffer; commit func([]byte) }
func (w *writeBuffer) Close() error { w.commit(w.Bytes()); return nil }

func (m *MockFS) Create(ctx context.Context, path string, overwrite bool) (io.WriteCloser, error) { return &writeBuffer{commit: func(b []byte){ m.put(path,b) }}, nil }

// Test helper
func (m *MockFS) PutTestFile(path string, data []byte) { m.put(path, data) }
