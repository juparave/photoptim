package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sftpCmd = &cobra.Command{
	Use:   "sftp",
	Short: "Run SFTP optimization workflow (TUI or batch)",
	RunE: func(cmd *cobra.Command, args []string) error {
		batch, _ := cmd.Flags().GetBool("batch")
		if batch {
			// Placeholder batch mode implementation
			fmt.Println("[SFTP batch] Not yet implemented")
			return nil
		}
		fmt.Println("[SFTP TUI] Not yet implemented")
		return nil
	},
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
