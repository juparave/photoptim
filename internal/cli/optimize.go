package cli

import (
	"fmt"

	"github.com/juparave/photoptim/internal/optimizer"

	"github.com/spf13/cobra"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize [input] [output]",
	Short: "Optimize an image",
	Long:  `Optimize an image file and save it to the specified output path.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outputPath := args[1]

		// Create optimizer
		opt := optimizer.New()

		// Get quality flag
		quality, err := cmd.Flags().GetInt("quality")
		if err != nil {
			return err
		}
		opt.Quality = quality

		// Optimize image
		if err := opt.Optimize(inputPath, outputPath); err != nil {
			return fmt.Errorf("optimization failed: %w", err)
		}

		fmt.Printf("Successfully optimized %s -> %s\n", inputPath, outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
	optimizeCmd.Flags().IntP("quality", "q", 80, "Quality for JPEG compression (1-100)")
}
