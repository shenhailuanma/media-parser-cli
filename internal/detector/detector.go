package detector

import (
	"fmt"
	"math"
	"strings"
)

type Problem struct {
	Severity    Severity          `json:"severity"`
	Category    Category          `json:"category"`
	Code        string            `json:"code"`
	Message     string            `json:"message"`
	Details     string            `json:"details,omitempty"`
	Suggestion  string            `json:"suggestion,omitempty"`
	Timestamp   float64           `json:"timestamp,omitempty"`
	StreamIndex int               `json:"stream_index,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityCritical
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	case SeverityError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Category int

const (
	CategoryCodec Category = iota
	CategoryContainer
	CategoryBitrate
	CategoryFrameRate
	CategoryResolution
	CategoryAudio
	CategoryTimestamp
	CategoryKeyframe
	CategoryPacketLoss
	CategoryCompatibility
)

func (c Category) String() string {
	switch c {
	case CategoryCodec:
		return "CODEC"
	case CategoryContainer:
		return "CONTAINER"
	case CategoryBitrate:
		return "BITRATE"
	case CategoryFrameRate:
		return "FRAMERATE"
	case CategoryResolution:
		return "RESOLUTION"
	case CategoryAudio:
		return "AUDIO"
	case CategoryTimestamp:
		return "TIMESTAMP"
	case CategoryKeyframe:
		return "KEYFRAME"
	case CategoryPacketLoss:
		return "PACKET_LOSS"
	case CategoryCompatibility:
		return "COMPATIBILITY"
	default:
		return "UNKNOWN"
	}
}

type Detector struct {
	problems []Problem
}

func New() *Detector {
	return &Detector{
		problems: make([]Problem, 0),
	}
}

func (d *Detector) GetProblems() []Problem {
	return d.problems
}

func (d *Detector) addProblem(problem Problem) {
	d.problems = append(d.problems, problem)
}

// DetectVideoProblems checks for common video stream issues
func (d *Detector) DetectVideoProblems(videoInfo interface{}) {
	// This will be populated with actual video info from analyzer
	// Placeholder for video problem detection logic
}

// DetectAudioProblems checks for common audio stream issues
func (d *Detector) DetectAudioProblems(audioInfo interface{}) {
	// Placeholder for audio problem detection logic
}

// DetectContainerProblems checks for container format issues
func (d *Detector) DetectContainerProblems(formatInfo interface{}) {
	// Placeholder for container problem detection logic
}

// DetectBitrateVariations analyzes bitrate consistency over time
func (d *Detector) DetectBitrateVariations(packets []PacketInfo) {
	if len(packets) < 2 {
		return
	}

	// Calculate average bitrate
	var totalBitrate float64
	var avgBitrate float64
	bitratePoints := make([]float64, 0)

	// Group packets by time window (1 second)
	timeWindow := 1.0
	currentWindow := 0.0
	windowBytes := 0

	for _, packet := range packets {
		if packet.PTS > currentWindow+timeWindow {
			if windowBytes > 0 {
				bitrate := float64(windowBytes) * 8 / timeWindow
				bitratePoints = append(bitratePoints, bitrate)
				totalBitrate += bitrate
			}
			currentWindow = packet.PTS
			windowBytes = packet.Size
		} else {
			windowBytes += packet.Size
		}
	}

	if len(bitratePoints) == 0 {
		return
	}

	avgBitrate = totalBitrate / float64(len(bitratePoints))

	// Calculate standard deviation
	var variance float64
	for _, bitrate := range bitratePoints {
		variance += math.Pow(bitrate-avgBitrate, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(bitratePoints)))

	// Check for high variance
	coefficientOfVariation := stdDev / avgBitrate
	if coefficientOfVariation > 0.3 {
		d.addProblem(Problem{
			Severity:   SeverityWarning,
			Category:   CategoryBitrate,
			Code:       "BITRATE_HIGH_VARIANCE",
			Message:    fmt.Sprintf("High bitrate variation detected (CV: %.2f)", coefficientOfVariation),
			Details:    fmt.Sprintf("Average: %.2f Mbps, StdDev: %.2f Mbps", avgBitrate/1000000, stdDev/1000000),
			Suggestion: "Consider using constant bitrate encoding or adjusting rate control settings",
		})
	}

	// Check for bitrate spikes
	for i, bitrate := range bitratePoints {
		if bitrate > avgBitrate*2.5 {
			d.addProblem(Problem{
				Severity:   SeverityWarning,
				Category:   CategoryBitrate,
				Code:       "BITRATE_SPIKE",
				Message:    fmt.Sprintf("Bitrate spike detected at ~%.2fs", float64(i)*timeWindow),
				Details:    fmt.Sprintf("Spike: %.2f Mbps (avg: %.2f Mbps)", bitrate/1000000, avgBitrate/1000000),
				Suggestion: "Review encoding settings or source content at this timestamp",
				Timestamp:  float64(i) * timeWindow,
			})
		}
	}
}

// DetectKeyframeIssues analyzes keyframe distribution
func (d *Detector) DetectKeyframeIssues(frames []FrameInfo) {
	if len(frames) == 0 {
		return
	}

	keyframes := make([]FrameInfo, 0)
	for _, frame := range frames {
		if frame.KeyFrame {
			keyframes = append(keyframes, frame)
		}
	}

	if len(keyframes) < 2 {
		d.addProblem(Problem{
			Severity:   SeverityError,
			Category:   CategoryKeyframe,
			Code:       "NO_KEYFRAMES",
			Message:    "No or insufficient keyframes detected",
			Suggestion: "Check encoder settings for keyframe interval",
		})
		return
	}

	// Calculate keyframe intervals
	intervals := make([]float64, 0)
	for i := 1; i < len(keyframes); i++ {
		interval := keyframes[i].PTS - keyframes[i-1].PTS
		intervals = append(intervals, interval)
	}

	// Calculate average interval
	var totalInterval float64
	for _, interval := range intervals {
		totalInterval += interval
	}
	avgInterval := totalInterval / float64(len(intervals))

	// Check for irregular keyframe intervals
	//for i, interval := range intervals {
	//	if math.Abs(interval-avgInterval)/avgInterval > 0.5 {
	//		d.addProblem(Problem{
	//			Severity:  SeverityWarning,
	//			Category:  CategoryKeyframe,
	//			Code:      "IRREGULAR_KEYFRAME_INTERVAL",
	//			Message:   fmt.Sprintf("Irregular keyframe interval at keyframe %d", i+1),
	//			Details:   fmt.Sprintf("Interval: %.2fs (expected: %.2fs)", interval, avgInterval),
	//			Timestamp: keyframes[i].PTS,
	//			Suggestion: "Consider using fixed GOP size for consistent keyframe intervals",
	//		})
	//	}
	//}

	// Check if keyframe interval is too large for streaming
	if avgInterval > 10.0 {
		d.addProblem(Problem{
			Severity:   SeverityWarning,
			Category:   CategoryKeyframe,
			Code:       "LARGE_KEYFRAME_INTERVAL",
			Message:    fmt.Sprintf("Large keyframe interval: %.2fs", avgInterval),
			Suggestion: "For streaming, consider reducing keyframe interval to 2-4 seconds",
		})
	}
}

// DetectTimestampIssues checks for PTS/DTS problems
func (d *Detector) DetectTimestampIssues(frames []FrameInfo) {
	if len(frames) < 2 {
		return
	}

	for i := 1; i < len(frames); i++ {
		// Check for non-monotonic PTS
		if frames[i].PTS < frames[i-1].PTS {
			d.addProblem(Problem{
				Severity:   SeverityError,
				Category:   CategoryTimestamp,
				Code:       "NON_MONOTONIC_PTS",
				Message:    fmt.Sprintf("Non-monotonic PTS at frame %d", i),
				Details:    fmt.Sprintf("Current: %.3f, Previous: %.3f", frames[i].PTS, frames[i-1].PTS),
				Timestamp:  frames[i].PTS,
				Suggestion: "Check source file or encoding process for timestamp issues",
			})
		}

		// Check for large PTS gaps
		ptsDiff := frames[i].PTS - frames[i-1].PTS
		if ptsDiff > 1.0 { // Gap larger than 1 second
			d.addProblem(Problem{
				Severity:   SeverityWarning,
				Category:   CategoryTimestamp,
				Code:       "LARGE_PTS_GAP",
				Message:    fmt.Sprintf("Large PTS gap at frame %d", i),
				Details:    fmt.Sprintf("Gap: %.3fs", ptsDiff),
				Timestamp:  frames[i].PTS,
				Suggestion: "Check for missing frames or timestamp discontinuities",
			})
		}

		// Check DTS order if available
		if frames[i].DTS > 0 && frames[i-1].DTS > 0 {
			if frames[i].DTS < frames[i-1].DTS {
				d.addProblem(Problem{
					Severity:   SeverityError,
					Category:   CategoryTimestamp,
					Code:       "NON_MONOTONIC_DTS",
					Message:    fmt.Sprintf("Non-monotonic DTS at frame %d", i),
					Details:    fmt.Sprintf("Current: %.3f, Previous: %.3f", frames[i].DTS, frames[i-1].DTS),
					Timestamp:  frames[i].PTS,
					Suggestion: "DTS must be monotonically increasing",
				})
			}

			// Check PTS >= DTS
			if frames[i].PTS < frames[i].DTS {
				d.addProblem(Problem{
					Severity:   SeverityError,
					Category:   CategoryTimestamp,
					Code:       "PTS_BEFORE_DTS",
					Message:    fmt.Sprintf("PTS before DTS at frame %d", i),
					Details:    fmt.Sprintf("PTS: %.3f, DTS: %.3f", frames[i].PTS, frames[i].DTS),
					Timestamp:  frames[i].PTS,
					Suggestion: "PTS must be greater than or equal to DTS",
				})
			}
		}
	}
}

// DetectPacketLoss analyzes for potential packet loss indicators
func (d *Detector) DetectPacketLoss(packets []PacketInfo) {
	if len(packets) < 2 {
		return
	}

	// Look for sudden PTS jumps that might indicate packet loss
	for i := 1; i < len(packets); i++ {
		ptsDiff := packets[i].PTS - packets[i-1].PTS

		// If PTS jump is larger than 0.5 seconds, might indicate loss
		if ptsDiff > 0.5 {
			d.addProblem(Problem{
				Severity:   SeverityWarning,
				Category:   CategoryPacketLoss,
				Code:       "POTENTIAL_PACKET_LOSS",
				Message:    fmt.Sprintf("Potential packet loss detected at %.2fs", packets[i].PTS),
				Details:    fmt.Sprintf("PTS jump of %.3fs detected", ptsDiff),
				Timestamp:  packets[i].PTS,
				Suggestion: "Check network conditions or source integrity",
			})
		}
	}
}

// AnalyzeCompatibility checks for compatibility issues
func (d *Detector) AnalyzeCompatibility(codec string, profile string, level int, container string) {
	// Check H.264 compatibility
	if strings.ToLower(codec) == "h264" {
		if strings.ToLower(profile) == "high" && level > 41 {
			d.addProblem(Problem{
				Severity:   SeverityWarning,
				Category:   CategoryCompatibility,
				Code:       "H264_COMPATIBILITY",
				Message:    fmt.Sprintf("H.264 High Profile Level %d.%d may have limited compatibility", level/10, level%10),
				Suggestion: "Consider using Main Profile Level 4.1 or lower for broader compatibility",
			})
		}
	}

	// Check HEVC compatibility
	if strings.ToLower(codec) == "hevc" || strings.ToLower(codec) == "h265" {
		d.addProblem(Problem{
			Severity:   SeverityInfo,
			Category:   CategoryCompatibility,
			Code:       "HEVC_SUPPORT",
			Message:    "HEVC/H.265 codec requires modern devices for playback",
			Suggestion: "Ensure target devices support HEVC or consider providing H.264 fallback",
		})
	}

	// Check container compatibility
	if strings.ToLower(container) == "mkv" || strings.ToLower(container) == "matroska" {
		d.addProblem(Problem{
			Severity:   SeverityInfo,
			Category:   CategoryCompatibility,
			Code:       "CONTAINER_COMPATIBILITY",
			Message:    "MKV container may have limited browser support",
			Suggestion: "Consider using MP4 container for web compatibility",
		})
	}
}

// PacketInfo represents a media packet
type PacketInfo struct {
	PTS         float64 `json:"pts"`
	DTS         float64 `json:"dts"`
	Size        int     `json:"size"`
	StreamIndex int     `json:"stream_index"`
	Flags       string  `json:"flags,omitempty"`
	Duration    float64 `json:"duration,omitempty"`
}

// FrameInfo represents a media frame
type FrameInfo struct {
	MediaType   string  `json:"media_type"`
	StreamIndex int     `json:"stream_index"`
	KeyFrame    bool    `json:"key_frame"`
	PTS         float64 `json:"pts"`
	DTS         float64 `json:"dts"`
	Duration    float64 `json:"duration"`
	Size        int     `json:"size"`
	PixFmt      string  `json:"pix_fmt,omitempty"`
	PictType    string  `json:"pict_type,omitempty"`
	CodedNumber int     `json:"coded_picture_number,omitempty"`
}

// BitratePoint represents a bitrate measurement at a specific time
type BitratePoint struct {
	Time    float64 `json:"time"`
	Bitrate float64 `json:"bitrate"`
	Type    string  `json:"type"` // "video", "audio", "total"
}

// GenerateBitrateTimeline creates bitrate timeline data from packets
func GenerateBitrateTimeline(packets []PacketInfo, windowSize float64) []BitratePoint {
	if len(packets) == 0 || windowSize <= 0 {
		return nil
	}

	points := make([]BitratePoint, 0)
	currentWindow := 0.0
	windowBytes := 0

	for _, packet := range packets {
		if packet.PTS > currentWindow+windowSize {
			if windowBytes > 0 {
				bitrate := float64(windowBytes) * 8 / windowSize
				points = append(points, BitratePoint{
					Time:    currentWindow,
					Bitrate: bitrate,
					Type:    "total",
				})
			}
			currentWindow += windowSize
			windowBytes = packet.Size
		} else {
			windowBytes += packet.Size
		}
	}

	// Add final window
	if windowBytes > 0 {
		bitrate := float64(windowBytes) * 8 / windowSize
		points = append(points, BitratePoint{
			Time:    currentWindow,
			Bitrate: bitrate,
			Type:    "total",
		})
	}

	return points
}
