package cli

import "testing"

func TestSFTPCmdRegistered(t *testing.T) {
    if _, _, err := rootCmd.Find([]string{"sftp"}); err != nil {
        t.Fatalf("sftp command not registered: %v", err)
    }
}
