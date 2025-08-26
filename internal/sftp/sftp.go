package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"time"

	pkgsftp "github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"

	"photoptim/internal/remotefs"
)

// Client implements remotefs.RemoteFS over SFTP.
type Client struct {
	cfg        remotefs.ConnectionConfig
	sshClient  *gossh.Client
	sftpClient *pkgsftp.Client
	root       string
}

// Connect establishes an SFTP session.
func (c *Client) Connect(ctx context.Context, cfg remotefs.ConnectionConfig) error {
	c.cfg = cfg
	if cfg.Port == 0 {
		cfg.Port = 22
	}

	authMethods, err := buildAuth(cfg)
	if err != nil {
		return err
	}
	sshConfig := &gossh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: c.hostKeyCallback(),
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, sshConfig)
	if err != nil {
		return fmt.Errorf("ssh handshake: %w", err)
	}
	c.sshClient = gossh.NewClient(sshConn, chans, reqs)

	s, err := pkgsftp.NewClient(c.sshClient)
	if err != nil {
		return fmt.Errorf("new sftp: %w", err)
	}
	c.sftpClient = s

	// Normalize root (ensure exists)
	root := cfg.RemotePath
	if root == "" {
		root = "/"
	}
	c.root = root
	return nil
}

func (c *Client) Close() error {
	if c.sftpClient != nil {
		_ = c.sftpClient.Close()
	}
	if c.sshClient != nil {
		_ = c.sshClient.Close()
	}
	return nil
}

func (c *Client) List(ctx context.Context, path string) ([]remotefs.RemoteEntry, error) {
	p := c.abs(path)
	fis, err := c.sftpClient.ReadDir(p)
	if err != nil {
		return nil, err
	}
	out := make([]remotefs.RemoteEntry, 0, len(fis))
	for _, fi := range fis {
		e := remotefs.RemoteEntry{
			Path:    filepath.Join(path, fi.Name()),
			Name:    fi.Name(),
			Size:    fi.Size(),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime(),
			IsDir:   fi.IsDir(),
		}
		out = append(out, e)
	}
	return out, nil
}

func (c *Client) Stat(ctx context.Context, path string) (remotefs.RemoteEntry, error) {
	fi, err := c.sftpClient.Stat(c.abs(path))
	if err != nil {
		return remotefs.RemoteEntry{}, err
	}
	return remotefs.RemoteEntry{Path: path, Name: filepath.Base(path), Size: fi.Size(), Mode: fi.Mode(), ModTime: fi.ModTime(), IsDir: fi.IsDir()}, nil
}

func (c *Client) Open(ctx context.Context, path string) ( /*nolint:ireturn*/ io.ReadCloser, remotefs.RemoteEntry, error) {
	f, err := c.sftpClient.Open(c.abs(path))
	if err != nil {
		return nil, remotefs.RemoteEntry{}, err
	}
	fi, _ := f.Stat()
	entry := remotefs.RemoteEntry{Path: path, Name: filepath.Base(path), Size: fi.Size(), Mode: fi.Mode(), ModTime: fi.ModTime(), IsDir: fi.IsDir()}
	return f, entry, nil
}

func (c *Client) Create(ctx context.Context, path string, overwrite bool) ( /*nolint:ireturn*/ io.WriteCloser, error) {
	full := c.abs(path)
	if !overwrite {
		if _, err := c.sftpClient.Stat(full); err == nil {
			return nil, errors.New("file exists")
		}
	}
	return c.sftpClient.Create(full)
}

func (c *Client) Join(elem ...string) string { return filepath.Join(elem...) }
func (c *Client) Root() string               { return c.root }

func (c *Client) abs(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Clean(filepath.Join(c.root, p))
}

// hostKeyCallback returns a callback that prompts the user (placeholder now always accepts).
func (c *Client) hostKeyCallback() gossh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		// TODO: implement fingerprint prompt + persistence
		return nil
	}
}

// buildAuth builds SSH auth methods (placeholder: password only if provided).
func buildAuth(cfg remotefs.ConnectionConfig) ([]gossh.AuthMethod, error) {
	methods := []gossh.AuthMethod{}
	// TODO: agent, key parsing, passphrase prompt
	if cfg.Password != "" {
		methods = append(methods, gossh.Password(cfg.Password))
	}
	if len(methods) == 0 {
		return nil, errors.New("no auth methods available (provide key or password)")
	}
	return methods, nil
}
