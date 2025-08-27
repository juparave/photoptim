package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"photoptim/internal/remotefs"
	sftpfs "photoptim/internal/sftp"
	"photoptim/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
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

		if batch {
			// batch mode validation
			missing := []string{}
			if host == "" {
				missing = append(missing, "--host")
			}
			if user == "" {
				missing = append(missing, "--user")
			}
			if len(missing) > 0 {
				return fmt.Errorf("missing required flags in batch mode: %s", strings.Join(missing, ", "))
			}

			// TODO: Implement full batch-mode pipeline here.
			// For now, we can leave the existing connectivity test as a placeholder.
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
			fmt.Println("SFTP basic connectivity verified. (Batch pipeline not yet wired here.)")

		} else {
			// Interactive TUI mode
			model := tui.NewSFTPModel()
			program := tea.NewProgram(&model)

			// Run the program
			if _, err := program.Run(); err != nil {
				return fmt.Errorf("error running program: %w", err)
			}
		}

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
