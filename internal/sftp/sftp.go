package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pkgsftp "github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/juparave/photoptim/internal/config"
	"github.com/juparave/photoptim/internal/remotefs"
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
		Timeout:         10 * time.Second,
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

	// Determine chroot: user-specified path OR user's home (working directory) by default
	wd, wdErr := s.Getwd()
	root := cfg.RemotePath
	if root == "" || root == "." {
		if wdErr == nil && wd != "" {
			root = wd
		} else {
			root = "/"
		}
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

	// Write to a temp file alongside the target, then atomically rename on
	// success. This protects the original from a partial/failed write.
	tmp := full + ".photoptim-tmp"
	f, err := c.sftpClient.Create(tmp)
	if err != nil {
		return nil, err
	}
	return &atomicWriteCloser{client: c.sftpClient, file: f, tmp: tmp, final: full}, nil
}

// atomicWriteCloser writes to a temp file and renames it over the final path on
// Close. Any error before a successful rename leaves the original file intact
// and removes the temp file.
type atomicWriteCloser struct {
	client *pkgsftp.Client
	file   *pkgsftp.File
	tmp    string
	final  string
	failed bool
}

func (w *atomicWriteCloser) Write(p []byte) (int, error) {
	n, err := w.file.Write(p)
	if err != nil {
		w.failed = true
	}
	return n, err
}

func (w *atomicWriteCloser) Close() error {
	closeErr := w.file.Close()
	if w.failed || closeErr != nil {
		_ = w.client.Remove(w.tmp)
		if closeErr != nil {
			return closeErr
		}
		return errors.New("write failed; original left unchanged")
	}

	// PosixRename atomically replaces the destination; fall back to a
	// remove+rename for servers without the posix-rename extension.
	if err := w.client.PosixRename(w.tmp, w.final); err != nil {
		if rmErr := w.client.Remove(w.final); rmErr == nil {
			if rnErr := w.client.Rename(w.tmp, w.final); rnErr == nil {
				return nil
			}
		}
		_ = w.client.Remove(w.tmp)
		return fmt.Errorf("rename temp to final: %w", err)
	}
	return nil
}

func (c *Client) Join(elem ...string) string { return filepath.Join(elem...) }
func (c *Client) Root() string               { return c.root }

func (c *Client) abs(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if p == "" || p == "." {
		return c.root
	}
	// Force relative to chroot
	if filepath.IsAbs(p) {
		p = strings.TrimPrefix(p, "/")
	}
	full := filepath.Join(c.root, p)
	clean := filepath.Clean(full)
	return clean
}

// hostKeyMu serializes appends to the known_hosts file across concurrent connects.
var hostKeyMu sync.Mutex

// hostKeyCallback verifies the server's host key against ~/.config/photoptim/known_hosts.
//
// On a known host it requires an exact match. On an unknown host it trusts on
// first use (TOFU): the key is persisted and accepted. A changed key (the MITM
// signal) is always rejected.
func (c *Client) hostKeyCallback() gossh.HostKeyCallback {
	path := config.ResolvePaths().KnownHosts

	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		hostKeyMu.Lock()
		defer hostKeyMu.Unlock()

		// Ensure the file exists so knownhosts.New can open it.
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(path, nil, 0o600); err != nil {
				return fmt.Errorf("create known_hosts: %w", err)
			}
		}

		check, err := knownhosts.New(path)
		if err != nil {
			return fmt.Errorf("load known_hosts: %w", err)
		}

		err = check(hostname, remote, key)
		if err == nil {
			return nil // known and matching
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			// Want entries present but none matched -> key changed (possible MITM).
			if len(keyErr.Want) > 0 {
				return fmt.Errorf("host key mismatch for %s: %w (remove the stale entry from %s if this change is expected)", hostname, err, path)
			}
			// No entries for this host -> trust on first use.
			if err := appendKnownHost(path, hostname, remote, key); err != nil {
				return fmt.Errorf("persist host key: %w", err)
			}
			return nil
		}

		return err
	}
}

// appendKnownHost appends an OpenSSH-format known_hosts line for the host.
func appendKnownHost(path, hostname string, remote net.Addr, key gossh.PublicKey) error {
	addrs := []string{hostname}
	if remote != nil {
		if ra := remote.String(); ra != "" && ra != hostname {
			addrs = append(addrs, ra)
		}
	}
	line := knownhosts.Line(addrs, key)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

// buildAuth builds SSH auth methods (password, ssh-agent, or identity files).
func buildAuth(cfg remotefs.ConnectionConfig) ([]gossh.AuthMethod, error) {
	methods := []gossh.AuthMethod{}

	// 1. SSH Agent Support (Highest priority, as it's the standard for seamless login)
	if agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		methods = append(methods, gossh.PublicKeysCallback(agent.NewClient(agentConn).Signers))
	}

	// 2. Identity Files Support
	home, err := os.UserHomeDir()
	if err == nil {
		keyPaths := []string{}
		// If user specified a key, try it first
		if cfg.KeyPath != "" {
			kp := cfg.KeyPath
			if strings.HasPrefix(kp, "~") {
				kp = filepath.Join(home, kp[1:])
			}
			keyPaths = append(keyPaths, kp)
		} else {
			// Otherwise try common defaults
			defaults := []string{
				filepath.Join(home, ".ssh", "id_ed25519"),
				filepath.Join(home, ".ssh", "id_rsa"),
				filepath.Join(home, ".ssh", "id_ecdsa"),
			}
			keyPaths = append(keyPaths, defaults...)
		}

		for _, kp := range keyPaths {
			if kp == "" {
				continue
			}
			key, err := os.ReadFile(kp)
			if err != nil {
				continue
			}

			signer, err := gossh.ParsePrivateKey(key)
			if err != nil {
				// If it's encrypted and we have a password, try using it as a passphrase
				if strings.Contains(err.Error(), "passphrase") && cfg.Password != "" {
					signer, err = gossh.ParsePrivateKeyWithPassphrase(key, []byte(cfg.Password))
				}
				if err != nil {
					continue
				}
			}
			methods = append(methods, gossh.PublicKeys(signer))
		}
	}

	// 3. Password Auth (Fallback)
	if cfg.Password != "" {
		methods = append(methods, gossh.Password(cfg.Password))
	}

	if len(methods) == 0 {
		return nil, errors.New("no auth methods available (ensure ssh-agent is running, default keys exist in ~/.ssh, or provide a password/key)")
	}
	return methods, nil
}
