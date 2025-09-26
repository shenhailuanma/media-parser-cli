package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tomi/media-parser-cli/internal/analyzer"
)

var (
	exportDir      string
	exportPackets  bool
	exportFrames   bool
	exportProblems bool
	exportBitrate  bool
	exportAll      bool
	maxPackets     int
	maxFrames      int
)

var exportCmd = &cobra.Command{
	Use:   "export [file or stream URL]",
	Short: "Export detailed media analysis to JSON files",
	Long: `Export analyzes a media file or stream and saves detailed information to JSON files.
	
The export command creates multiple JSON files in the specified directory:
- media_info.json: Basic media information
- problems.json: Detected issues and warnings
- packets.json: Packet-level data (optional)
- frames.json: Frame-level data (optional)
- bitrate_timeline.json: Bitrate over time (optional)

This is useful for:
- Detailed debugging and analysis
- Creating reports for quality assurance
- Visualizing data with external tools
- Archiving media analysis results

Examples:
  media-parser-cli export video.mp4 -d ./analysis
  media-parser-cli export stream.m3u8 -d ./reports --export-all
  media-parser-cli export video.mp4 -d ./debug --export-frames --max-frames 1000`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportDir, "dir", "d", "./media-analysis", "Directory to save analysis files")
	exportCmd.Flags().BoolVar(&exportPackets, "export-packets", false, "Export packet information")
	exportCmd.Flags().BoolVar(&exportFrames, "export-frames", false, "Export frame information")
	exportCmd.Flags().BoolVar(&exportProblems, "export-problems", true, "Export detected problems")
	exportCmd.Flags().BoolVar(&exportBitrate, "export-bitrate", false, "Export bitrate timeline")
	exportCmd.Flags().BoolVar(&exportAll, "export-all", false, "Export all available information")
	exportCmd.Flags().IntVar(&maxPackets, "max-packets", 10000, "Maximum number of packets to export")
	exportCmd.Flags().IntVar(&maxFrames, "max-frames", 5000, "Maximum number of frames to export")
}

func runExport(cmd *cobra.Command, args []string) error {
	input := args[0]

	if exportAll {
		exportPackets = true
		exportFrames = true
		exportProblems = true
		exportBitrate = true
	}

	// Create export directory
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Exporting analysis to: %s\n", exportDir)
	}

	// Create timestamp for this export
	timestamp := time.Now().Format("20060102_150405")
	exportSubDir := filepath.Join(exportDir, fmt.Sprintf("analysis_%s", timestamp))
	if err := os.MkdirAll(exportSubDir, 0755); err != nil {
		return fmt.Errorf("failed to create export subdirectory: %w", err)
	}

	// Analyze media
	options := analyzer.Options{
		Timeout:       timeout,
		ShowVideo:     true,
		ShowAudio:     true,
		ShowFormat:    true,
		ShowStreams:   true,
		Verbose:       verbose,
		AnalyzePackets: exportPackets,
		AnalyzeFrames:  exportFrames,
		MaxPackets:    maxPackets,
		MaxFrames:     maxFrames,
	}

	analyzer := analyzer.New(options)
	result, err := analyzer.AnalyzeWithDetails(input)
	if err != nil {
		return fmt.Errorf("failed to analyze media: %w", err)
	}

	// Export basic media info
	if err := exportJSON(filepath.Join(exportSubDir, "media_info.json"), result.MediaInfo); err != nil {
		return fmt.Errorf("failed to export media info: %w", err)
	}
	fmt.Printf("✓ Exported media info to %s\n", filepath.Join(exportSubDir, "media_info.json"))

	// Export problems
	if exportProblems && len(result.Problems) > 0 {
		if err := exportJSON(filepath.Join(exportSubDir, "problems.json"), result.Problems); err != nil {
			return fmt.Errorf("failed to export problems: %w", err)
		}
		fmt.Printf("✓ Exported %d problems to %s\n", len(result.Problems), filepath.Join(exportSubDir, "problems.json"))
	}

	// Export packets
	if exportPackets && len(result.Packets) > 0 {
		if err := exportJSON(filepath.Join(exportSubDir, "packets.json"), result.Packets); err != nil {
			return fmt.Errorf("failed to export packets: %w", err)
		}
		fmt.Printf("✓ Exported %d packets to %s\n", len(result.Packets), filepath.Join(exportSubDir, "packets.json"))
	}

	// Export frames
	if exportFrames && len(result.Frames) > 0 {
		if err := exportJSON(filepath.Join(exportSubDir, "frames.json"), result.Frames); err != nil {
			return fmt.Errorf("failed to export frames: %w", err)
		}
		fmt.Printf("✓ Exported %d frames to %s\n", len(result.Frames), filepath.Join(exportSubDir, "frames.json"))

		// Export frame visualization (eyecard-style)
		if frameViz := generateFrameVisualization(result.Frames); frameViz != nil {
			if err := exportJSON(filepath.Join(exportSubDir, "frame_visualization.json"), frameViz); err != nil {
				return fmt.Errorf("failed to export frame visualization: %w", err)
			}
			fmt.Printf("✓ Exported frame visualization to %s\n", filepath.Join(exportSubDir, "frame_visualization.json"))
		}
	}

	// Export bitrate timeline
	if exportBitrate && len(result.BitrateTimeline) > 0 {
		if err := exportJSON(filepath.Join(exportSubDir, "bitrate_timeline.json"), result.BitrateTimeline); err != nil {
			return fmt.Errorf("failed to export bitrate timeline: %w", err)
		}
		fmt.Printf("✓ Exported bitrate timeline to %s\n", filepath.Join(exportSubDir, "bitrate_timeline.json"))
	}

	// Create summary file
	summary := map[string]interface{}{
		"analysis_timestamp": timestamp,
		"input_file":        input,
		"export_directory":  exportSubDir,
		"files_created": map[string]bool{
			"media_info.json":         true,
			"problems.json":          exportProblems && len(result.Problems) > 0,
			"packets.json":           exportPackets && len(result.Packets) > 0,
			"frames.json":            exportFrames && len(result.Frames) > 0,
			"frame_visualization.json": exportFrames && len(result.Frames) > 0,
			"bitrate_timeline.json":  exportBitrate && len(result.BitrateTimeline) > 0,
		},
		"statistics": map[string]int{
			"problems_found": len(result.Problems),
			"packets_analyzed": len(result.Packets),
			"frames_analyzed": len(result.Frames),
		},
	}

	if err := exportJSON(filepath.Join(exportSubDir, "summary.json"), summary); err != nil {
		return fmt.Errorf("failed to export summary: %w", err)
	}

	fmt.Printf("\n")
	fmt.Printf("Analysis exported to: %s\n", exportSubDir)
	fmt.Printf("Total files created: %d\n", countCreatedFiles(summary["files_created"].(map[string]bool)))

	return nil
}

func exportJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func countCreatedFiles(files map[string]bool) int {
	count := 0
	for _, created := range files {
		if created {
			count++
		}
	}
	return count
}

type FrameVisualization struct {
	TotalFrames int                    `json:"total_frames"`
	Duration    float64                `json:"duration"`
	FrameTypes  map[string]int         `json:"frame_types"`
	GOPStructure []GOPInfo             `json:"gop_structure"`
	Timeline    []FrameTimelineEntry   `json:"timeline"`
}

type GOPInfo struct {
	StartTime   float64 `json:"start_time"`
	EndTime     float64 `json:"end_time"`
	FrameCount  int     `json:"frame_count"`
	IFrames     int     `json:"i_frames"`
	PFrames     int     `json:"p_frames"`
	BFrames     int     `json:"b_frames"`
}

type FrameTimelineEntry struct {
	Time      float64 `json:"time"`
	FrameType string  `json:"frame_type"`
	Size      int     `json:"size"`
	KeyFrame  bool    `json:"key_frame"`
}

func generateFrameVisualization(frames []analyzer.FrameData) *FrameVisualization {
	if len(frames) == 0 {
		return nil
	}

	viz := &FrameVisualization{
		TotalFrames: len(frames),
		FrameTypes:  make(map[string]int),
		GOPStructure: make([]GOPInfo, 0),
		Timeline:    make([]FrameTimelineEntry, 0),
	}

	// Find video frames only
	videoFrames := make([]analyzer.FrameData, 0)
	for _, frame := range frames {
		if strings.ToLower(frame.MediaType) == "video" {
			videoFrames = append(videoFrames, frame)
		}
	}

	if len(videoFrames) == 0 {
		return viz
	}

	// Count frame types and build timeline
	currentGOP := GOPInfo{StartTime: videoFrames[0].PTS}
	
	for i, frame := range videoFrames {
		// Count frame types
		if frame.PictType != "" {
			viz.FrameTypes[frame.PictType]++
			
			switch frame.PictType {
			case "I":
				currentGOP.IFrames++
			case "P":
				currentGOP.PFrames++
			case "B":
				currentGOP.BFrames++
			}
		}

		// Add to timeline (sample every N frames for large files)
		sampleRate := len(videoFrames) / 1000 // Max 1000 timeline entries
		if sampleRate < 1 {
			sampleRate = 1
		}
		if i%sampleRate == 0 {
			viz.Timeline = append(viz.Timeline, FrameTimelineEntry{
				Time:      frame.PTS,
				FrameType: frame.PictType,
				Size:      frame.Size,
				KeyFrame:  frame.KeyFrame,
			})
		}

		// Detect GOP boundaries (I-frame starts new GOP)
		if frame.KeyFrame && i > 0 {
			currentGOP.EndTime = videoFrames[i-1].PTS
			currentGOP.FrameCount = currentGOP.IFrames + currentGOP.PFrames + currentGOP.BFrames
			viz.GOPStructure = append(viz.GOPStructure, currentGOP)
			currentGOP = GOPInfo{StartTime: frame.PTS, IFrames: 1}
		} else {
			currentGOP.FrameCount++
		}
	}

	// Add final GOP
	if currentGOP.FrameCount > 0 {
		currentGOP.EndTime = videoFrames[len(videoFrames)-1].PTS
		viz.GOPStructure = append(viz.GOPStructure, currentGOP)
	}

	viz.Duration = videoFrames[len(videoFrames)-1].PTS

	return viz
}