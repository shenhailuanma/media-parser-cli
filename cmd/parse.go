package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tomi/media-parser-cli/internal/analyzer"
	"github.com/tomi/media-parser-cli/internal/reporter"
)

var (
	showVideo   bool
	showAudio   bool
	showFormat  bool
	showStreams bool
	showAll     bool
	timeout     int
)

var parseCmd = &cobra.Command{
	Use:   "parse [file or stream URL]",
	Short: "Parse and analyze a video file or stream",
	Long: `Parse analyzes a video file or stream URL and provides detailed media information.
	
Supported inputs:
- Local video files (mp4, mkv, avi, mov, etc.)
- HTTP/HTTPS streams (HLS, DASH, direct media URLs)
- RTMP/RTSP streams
- Network file paths

Examples:
  media-parser-cli parse video.mp4
  media-parser-cli parse https://example.com/stream.m3u8
  media-parser-cli parse rtmp://server/live/stream
  media-parser-cli parse --show-all video.mp4 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().BoolVar(&showVideo, "show-video", true, "Show video stream information")
	parseCmd.Flags().BoolVar(&showAudio, "show-audio", true, "Show audio stream information")
	parseCmd.Flags().BoolVar(&showFormat, "show-format", true, "Show container format information")
	parseCmd.Flags().BoolVar(&showStreams, "show-streams", false, "Show all stream details")
	parseCmd.Flags().BoolVar(&showAll, "show-all", false, "Show all available information")
	parseCmd.Flags().IntVar(&timeout, "timeout", 30, "Analysis timeout in seconds")
}

func runParse(cmd *cobra.Command, args []string) error {
	input := args[0]

	if showAll {
		showVideo = true
		showAudio = true
		showFormat = true
		showStreams = true
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Analyzing: %s\n", input)
	}

	options := analyzer.Options{
		Timeout:     timeout,
		ShowVideo:   showVideo,
		ShowAudio:   showAudio,
		ShowFormat:  showFormat,
		ShowStreams: showStreams,
		Verbose:     verbose,
	}

	analyzer := analyzer.New(options)
	result, err := analyzer.Analyze(input)
	if err != nil {
		return fmt.Errorf("failed to analyze media: %w", err)
	}

	reporterOptions := reporter.Options{
		Format:  getOutputFormat(),
		Verbose: verbose,
	}

	reporter := reporter.New(reporterOptions)
	if err := reporter.Print(result); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	return nil
}

func getOutputFormat() reporter.Format {
	switch strings.ToLower(output) {
	case "json":
		return reporter.FormatJSON
	case "yaml":
		return reporter.FormatYAML
	case "text", "":
		return reporter.FormatText
	default:
		fmt.Fprintf(os.Stderr, "Unknown output format: %s, using text\n", output)
		return reporter.FormatText
	}
}