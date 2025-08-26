package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"photoptim/internal/remotefs"
	sftpfs "photoptim/internal/sftp"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var sftpCmd = &cobra.Command{
	Use:   "sftp",
	Short: "Run SFTP optimization workflow (TUI or batch)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Collect flags
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		remotePath, _ := cmd.Flags().GetString("remote-path")
		keyPath, _ := cmd.Flags().GetString("key")
		password, _ := cmd.Flags().GetString("password")
		batch, _ := cmd.Flags().GetBool("batch")

		if port == 0 {
			port = 22
		}

		if !batch { // interactive prompting
			rdr := bufio.NewReader(os.Stdin)
			if host == "" {
				host = prompt(rdr, "Host: ")
			}
			if pStr := promptDefault(rdr, fmt.Sprintf("Port [%d]: ", port), strconv.Itoa(port)); pStr != "" {
				if v, err := strconv.Atoi(pStr); err == nil {
					port = v
				}
			}
			if user == "" {
				user = prompt(rdr, "User: ")
			}
			if remotePath == "" { // let user specify or accept default '/'
				remotePath = promptDefault(rdr, "Remote path [/] (use / for home): ", "/")
			}
			if keyPath == "" && password == "" { // ask for password only if neither provided
				if term.IsTerminal(int(os.Stdin.Fd())) {
					fmt.Print("Password (leave empty to skip password auth): ")
					b, _ := term.ReadPassword(int(os.Stdin.Fd()))
					fmt.Println()
					password = strings.TrimSpace(string(b))
				} else {
					password = strings.TrimSpace(prompt(rdr, "Password (leave empty to skip password auth): "))
				}
			}
		} else { // batch mode validation
			missing := []string{}
			if host == "" {
				missing = append(missing, "--host")
			}
			if user == "" {
				missing = append(missing, "--user")
			}
			if len(missing) > 0 { // remote-path optional in batch
				return fmt.Errorf("missing required flags in batch mode: %s", strings.Join(missing, ", "))
			}
		}

		// Treat '/' or empty remotePath as request to use user home as chroot (handled in client)
		if remotePath == "/" {
			remotePath = ""
		}

		fmt.Printf("Connecting to %s@%s:%d (path=%s) ...\n", user, host, port, func() string {
			if remotePath == "" {
				return "<home>"
			}
			return remotePath
		}())
		cfg := remotefs.ConnectionConfig{Host: host, Port: port, User: user, Password: password, KeyPath: keyPath, RemotePath: remotePath}
		client := &sftpfs.Client{}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Connect(ctx, cfg); err != nil {
			return fmt.Errorf("sftp connect failed: %w", err)
		}
		defer client.Close()
		// List root
		entries, err := client.List(ctx, ".")
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		if len(entries) == 0 {
			fmt.Println("(remote directory is empty)")
		} else {
			fmt.Println("Remote entries:")
			for _, e := range entries {
				typ := "file"
				if e.IsDir {
					typ = "dir"
				}
				fmt.Printf("  %-30s %7d %s\n", e.Name, e.Size, typ)
			}
		}
		fmt.Println("SFTP basic connectivity verified. (Optimization pipeline not yet wired here.)")
		if batch {
			return nil
		}
		fmt.Println("Future: launch SFTP TUI states (connection/browser/selection).")
		return nil
	},
}

func prompt(r *bufio.Reader, label string) string {
	fmt.Print(label)
	txt, _ := r.ReadString('\n')
	return strings.TrimSpace(txt)
}
func promptDefault(r *bufio.Reader, label, def string) string {
	fmt.Print(label)
	txt, _ := r.ReadString('\n')
	txt = strings.TrimSpace(txt)
	if txt == "" {
		return def
	}
	return txt
}

func init() {
	rootCmd.AddCommand(sftpCmd)
	sftpCmd.Flags().String("host", "", "SFTP host")
	sftpCmd.Flags().Int("port", 22, "SFTP port")
	sftpCmd.Flags().String("user", "", "Username")
	sftpCmd.Flags().String("remote-path", "/", "Initial remote path (chroot)")
	sftpCmd.Flags().String("key", "", "Private key path")
	sftpCmd.Flags().String("password", "", "Password (fallback)")
	sftpCmd.Flags().Int("quality", 80, "JPEG quality (1-100)")
	sftpCmd.Flags().String("size-threshold", "", "Inclusive size threshold e.g. 500KB, 2MB")
	sftpCmd.Flags().Int("concurrency", 4, "Worker concurrency")
	sftpCmd.Flags().String("ttl", "2m", "Directory cache TTL")
	sftpCmd.Flags().Bool("save-config", false, "Persist settings to config file")
	sftpCmd.Flags().Bool("verbose", false, "Verbose logging")
	sftpCmd.Flags().Bool("audit", false, "Enable audit logging")
	sftpCmd.Flags().Bool("skip-cache", false, "Skip directory cache")
	sftpCmd.Flags().Bool("keep-temp", false, "Keep temp files after completion")
	sftpCmd.Flags().Bool("batch", false, "Run in non-interactive batch mode")
}
