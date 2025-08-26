package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"photoptim/internal/optimizer"

	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch [input directory] [output directory]",
	Short: "Optimize multiple images",
	Long:  `Optimize all images in a directory and save them to the output directory.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir := args[0]
		outputDir := args[1]

		// Create optimizer
		opt := optimizer.New()

		// Get quality flag
		quality, err := cmd.Flags().GetInt("quality")
		if err != nil {
			return err
		}
		opt.Quality = quality

		// Read all files in input directory
		files, err := filepath.Glob(filepath.Join(inputDir, "*"))
		if err != nil {
			return fmt.Errorf("failed to read input directory: %w", err)
		}

		// Process each file
		count := 0
		for _, file := range files {
			// Check if it's an image file
			ext := strings.ToLower(filepath.Ext(file))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				// Generate output path
				filename := filepath.Base(file)
				outputPath := filepath.Join(outputDir, filename)

				// Optimize image
				if err := opt.Optimize(file, outputPath); err != nil {
					fmt.Printf("Warning: failed to optimize %s: %v\n", file, err)
					continue
				}
				count++
			}
		}

		fmt.Printf("Successfully optimized %d images\n", count)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)
	batchCmd.Flags().IntP("quality", "q", 80, "Quality for JPEG compression (1-100)")
}
