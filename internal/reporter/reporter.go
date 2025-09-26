package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tomi/media-parser-cli/internal/analyzer"
	"github.com/tomi/media-parser-cli/internal/detector"
)

type Format int

const (
	FormatText Format = iota
	FormatJSON
	FormatYAML
)

type Options struct {
	Format       Format
	Verbose      bool
	ShowProblems bool
}

type Reporter struct {
	options Options
	writer  io.Writer
}

func New(options Options) *Reporter {
	return &Reporter{
		options: options,
		writer:  os.Stdout,
	}
}

func (r *Reporter) Print(info *analyzer.MediaInfo) error {
	switch r.options.Format {
	case FormatJSON:
		return r.printJSON(info)
	case FormatYAML:
		return r.printYAML(info)
	case FormatText:
		return r.printText(info)
	default:
		return r.printText(info)
	}
}

func (r *Reporter) printJSON(info *analyzer.MediaInfo) error {
	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func (r *Reporter) printYAML(info *analyzer.MediaInfo) error {
	jsonData, err := json.Marshal(info)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	return r.printYAMLMap(data, 0)
}

func (r *Reporter) printYAMLMap(data interface{}, indent int) error {
	indentStr := strings.Repeat("  ", indent)

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if value == nil {
				continue
			}
			fmt.Fprintf(r.writer, "%s%s:", indentStr, key)
			if mapVal, ok := value.(map[string]interface{}); ok && len(mapVal) > 0 {
				fmt.Fprintln(r.writer)
				r.printYAMLMap(value, indent+1)
			} else if a, ok := value.([]interface{}); ok && len(a) > 0 {
				fmt.Fprintln(r.writer)
				for _, item := range a {
					fmt.Fprintf(r.writer, "%s- ", strings.Repeat("  ", indent+1))
					if _, ok := item.(map[string]interface{}); ok {
						fmt.Fprintln(r.writer)
						r.printYAMLMap(item, indent+2)
					} else {
						fmt.Fprintln(r.writer, item)
					}
				}
			} else {
				fmt.Fprintf(r.writer, " %v\n", value)
			}
		}
	default:
		fmt.Fprintf(r.writer, "%s%v\n", indentStr, v)
	}

	return nil
}

func (r *Reporter) printText(info *analyzer.MediaInfo) error {
	fmt.Fprintln(r.writer, strings.Repeat("=", 80))
	fmt.Fprintf(r.writer, "MEDIA ANALYSIS REPORT\n")
	fmt.Fprintf(r.writer, "Analyzed at: %s\n", info.AnalyzedAt.Format(time.RFC3339))
	fmt.Fprintf(r.writer, "Input: %s\n", info.Input)
	fmt.Fprintln(r.writer, strings.Repeat("=", 80))

	if info.Format != nil {
		fmt.Fprintln(r.writer, "\nCONTAINER FORMAT:")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		r.printFormatInfo(info.Format)
	}

	if info.VideoStream != nil {
		fmt.Fprintln(r.writer, "\nVIDEO STREAM:")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		r.printVideoInfo(info.VideoStream)
	}

	if info.AudioStream != nil {
		fmt.Fprintln(r.writer, "\nAUDIO STREAM:")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		r.printAudioInfo(info.AudioStream)
	}

	if len(info.Streams) > 0 {
		fmt.Fprintln(r.writer, "\nALL STREAMS:")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		r.printStreamsTable(info.Streams)
	}

	fmt.Fprintln(r.writer, strings.Repeat("=", 80))
	return nil
}

func (r *Reporter) printFormatInfo(format *analyzer.FormatInfo) {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Format:\t%s\n", format.FormatName)
	fmt.Fprintf(w, "Long Name:\t%s\n", format.FormatLongName)
	if format.Duration > 0 {
		fmt.Fprintf(w, "Duration:\t%s\n", r.formatDuration(format.Duration))
	}
	if format.Size > 0 {
		fmt.Fprintf(w, "File Size:\t%s\n", r.formatSize(format.Size))
	}
	if format.Bitrate > 0 {
		fmt.Fprintf(w, "Overall Bitrate:\t%s\n", r.formatBitrate(format.Bitrate))
	}
	if r.options.Verbose && format.ProbeScore > 0 {
		fmt.Fprintf(w, "Probe Score:\t%d\n", format.ProbeScore)
	}
	w.Flush()
}

func (r *Reporter) printVideoInfo(video *analyzer.VideoInfo) {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Stream Index:\t%d\n", video.Index)
	fmt.Fprintf(w, "Codec:\t%s (%s)\n", video.Codec, video.CodecLongName)
	if video.Profile != "" {
		fmt.Fprintf(w, "Profile:\t%s\n", video.Profile)
	}
	fmt.Fprintf(w, "Resolution:\t%dx%d\n", video.Width, video.Height)
	if video.AspectRatio != "" {
		fmt.Fprintf(w, "Aspect Ratio:\t%s\n", video.AspectRatio)
	}
	fmt.Fprintf(w, "Pixel Format:\t%s\n", video.PixelFormat)
	fmt.Fprintf(w, "Frame Rate:\t%s fps\n", video.FrameRate)
	if video.AvgFrameRate != "" && video.AvgFrameRate != video.FrameRate {
		fmt.Fprintf(w, "Avg Frame Rate:\t%s fps\n", video.AvgFrameRate)
	}
	if video.Bitrate > 0 {
		fmt.Fprintf(w, "Bitrate:\t%s\n", r.formatBitrate(video.Bitrate))
	}
	if video.Duration > 0 {
		fmt.Fprintf(w, "Duration:\t%s\n", r.formatDuration(video.Duration))
	}
	if video.FrameCount > 0 {
		fmt.Fprintf(w, "Total Frames:\t%d\n", video.FrameCount)
	}
	if r.options.Verbose {
		if video.ColorSpace != "" {
			fmt.Fprintf(w, "Color Space:\t%s\n", video.ColorSpace)
		}
		if video.ColorPrimaries != "" {
			fmt.Fprintf(w, "Color Primaries:\t%s\n", video.ColorPrimaries)
		}
		if video.ColorTransfer != "" {
			fmt.Fprintf(w, "Color Transfer:\t%s\n", video.ColorTransfer)
		}
		if video.HasBFrames > 0 {
			fmt.Fprintf(w, "Has B-Frames:\t%d\n", video.HasBFrames)
		}
	}
	w.Flush()
}

func (r *Reporter) printAudioInfo(audio *analyzer.AudioInfo) {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Stream Index:\t%d\n", audio.Index)
	fmt.Fprintf(w, "Codec:\t%s (%s)\n", audio.Codec, audio.CodecLongName)
	if audio.Profile != "" {
		fmt.Fprintf(w, "Profile:\t%s\n", audio.Profile)
	}
	fmt.Fprintf(w, "Channels:\t%d\n", audio.Channels)
	if audio.ChannelLayout != "" {
		fmt.Fprintf(w, "Channel Layout:\t%s\n", audio.ChannelLayout)
	}
	fmt.Fprintf(w, "Sample Rate:\t%d Hz\n", audio.SampleRate)
	fmt.Fprintf(w, "Sample Format:\t%s\n", audio.SampleFormat)
	if audio.Bitrate > 0 {
		fmt.Fprintf(w, "Bitrate:\t%s\n", r.formatBitrate(audio.Bitrate))
	}
	if audio.Duration > 0 {
		fmt.Fprintf(w, "Duration:\t%s\n", r.formatDuration(audio.Duration))
	}
	w.Flush()
}

func (r *Reporter) printStreamsTable(streams []analyzer.StreamInfo) {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Index\tType\tCodec\tTags\n")
	fmt.Fprintf(w, "-----\t----\t-----\t----\n")
	for _, stream := range streams {
		tags := ""
		if len(stream.Tags) > 0 {
			var tagPairs []string
			for k, v := range stream.Tags {
				tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", k, v))
			}
			tags = strings.Join(tagPairs, ", ")
			if len(tags) > 50 {
				tags = tags[:47] + "..."
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", stream.Index, stream.Type, stream.Codec, tags)
	}
	w.Flush()
}

func (r *Reporter) formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := duration.Seconds() - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, secs)
}

func (r *Reporter) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (r *Reporter) formatBitrate(bps int64) string {
	if bps < 1000 {
		return fmt.Sprintf("%d bps", bps)
	} else if bps < 1000000 {
		return fmt.Sprintf("%.1f kbps", float64(bps)/1000)
	} else {
		return fmt.Sprintf("%.2f Mbps", float64(bps)/1000000)
	}
}

// PrintDetailed prints detailed analysis including problems
func (r *Reporter) PrintDetailed(analysis *analyzer.DetailedAnalysis) error {
	switch r.options.Format {
	case FormatJSON:
		return r.printDetailedJSON(analysis)
	case FormatYAML:
		return r.printDetailedYAML(analysis)
	case FormatText:
		return r.printDetailedText(analysis)
	default:
		return r.printDetailedText(analysis)
	}
}

func (r *Reporter) printDetailedJSON(analysis *analyzer.DetailedAnalysis) error {
	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func (r *Reporter) printDetailedYAML(analysis *analyzer.DetailedAnalysis) error {
	jsonData, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	return r.printYAMLMap(data, 0)
}

func (r *Reporter) printDetailedText(analysis *analyzer.DetailedAnalysis) error {
	// First print the basic media info
	if err := r.printText(analysis.MediaInfo); err != nil {
		return err
	}

	// Then print detected problems
	if r.options.ShowProblems && len(analysis.Problems) > 0 {
		fmt.Fprintln(r.writer, "\nDETECTED PROBLEMS:")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		r.printProblems(analysis.Problems)
	}

	return nil
}

func (r *Reporter) printProblems(problems []detector.Problem) {
	// Group problems by severity
	var errors, criticals, warnings, infos []detector.Problem

	for _, p := range problems {
		switch p.Severity {
		case detector.SeverityError:
			errors = append(errors, p)
		case detector.SeverityCritical:
			criticals = append(criticals, p)
		case detector.SeverityWarning:
			warnings = append(warnings, p)
		case detector.SeverityInfo:
			infos = append(infos, p)
		}
	}

	// Print errors first
	if len(errors) > 0 {
		fmt.Fprintln(r.writer, "\n🔴 ERRORS:")
		for _, p := range errors {
			r.printProblem(p)
		}
	}

	// Then critical issues
	if len(criticals) > 0 {
		fmt.Fprintln(r.writer, "\n🟠 CRITICAL:")
		for _, p := range criticals {
			r.printProblem(p)
		}
	}

	// Then warnings
	if len(warnings) > 0 {
		fmt.Fprintln(r.writer, "\n🟡 WARNINGS:")
		for _, p := range warnings {
			r.printProblem(p)
		}
	}

	// Finally info
	if len(infos) > 0 && r.options.Verbose {
		fmt.Fprintln(r.writer, "\n🔵 INFO:")
		for _, p := range infos {
			r.printProblem(p)
		}
	}

	// Summary
	fmt.Fprintln(r.writer, "\n" + strings.Repeat("-", 40))
	fmt.Fprintf(r.writer, "Summary: %d errors, %d critical, %d warnings, %d info\n",
		len(errors), len(criticals), len(warnings), len(infos))
}

func (r *Reporter) printProblem(p detector.Problem) {
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	
	fmt.Fprintf(w, "  [%s]\t%s\n", p.Code, p.Message)
	
	if p.Details != "" {
		fmt.Fprintf(w, "    Details:\t%s\n", p.Details)
	}
	
	if p.Suggestion != "" {
		fmt.Fprintf(w, "    ✨ Suggestion:\t%s\n", p.Suggestion)
	}
	
	if p.Timestamp > 0 {
		fmt.Fprintf(w, "    Timestamp:\t%.2fs\n", p.Timestamp)
	}
	
	w.Flush()
	fmt.Fprintln(r.writer)
}