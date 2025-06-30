package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation for cu",
	Long:  `Generate documentation for cu in various formats including Markdown, Man pages, and RST.`,
	Hidden: true, // Hide from regular help output
}

var genMarkdownCmd = &cobra.Command{
	Use:   "markdown",
	Short: "Generate Markdown documentation",
	Long:  `Generate Markdown documentation for all cu commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "./docs"
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Generate markdown documentation
		if err := doc.GenMarkdownTree(rootCmd, dir); err != nil {
			return fmt.Errorf("failed to generate markdown docs: %w", err)
		}

		fmt.Printf("Documentation generated in %s\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.AddCommand(genMarkdownCmd)
	
	genMarkdownCmd.Flags().StringP("dir", "d", "./docs", "Directory to write documentation files")
}