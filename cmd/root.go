package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	verbose bool
	output  string
)

var rootCmd = &cobra.Command{
	Use:   "media-parser-cli",
	Short: "A CLI tool for analyzing video files and streams",
	Long: `Media Parser CLI is a command-line tool for analyzing video files and streams.
	
It provides detailed media information including:
- Video codec, resolution, framerate, bitrate
- Audio codec, channels, sample rate, bitrate
- Container format information
- Stream metadata
- Timing and duration information

Perfect for debugging media issues and understanding media properties.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format (json, yaml, text)")
}