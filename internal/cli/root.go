package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "photoptim",
	Short: "Photoptim is a tool for optimizing images",
	Long: `Photoptim is a fast and flexible tool for optimizing images.
It supports various formats and optimization techniques.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Photoptim!")
		fmt.Println("Use 'photoptim --help' for more information.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}