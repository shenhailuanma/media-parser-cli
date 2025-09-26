package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/tomi/media-parser-cli/internal/detector"
	"github.com/tomi/media-parser-cli/pkg/ffprobe"
)

type Options struct {
	Timeout        int
	ShowVideo      bool
	ShowAudio      bool
	ShowFormat     bool
	ShowStreams    bool
	Verbose        bool
	AnalyzePackets bool
	AnalyzeFrames  bool
	MaxPackets     int
	MaxFrames      int
}

type Analyzer struct {
	options Options
	ffprobe *ffprobe.FFProbe
}

type MediaInfo struct {
	Input       string         `json:"input"`
	Format      *FormatInfo    `json:"format,omitempty"`
	VideoStream *VideoInfo     `json:"video,omitempty"`
	AudioStream *AudioInfo     `json:"audio,omitempty"`
	Streams     []StreamInfo   `json:"streams,omitempty"`
	AnalyzedAt  time.Time      `json:"analyzed_at"`
}

type FormatInfo struct {
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	Duration       float64           `json:"duration"`
	Size           int64             `json:"size"`
	Bitrate        int64             `json:"bitrate"`
	ProbeScore     int               `json:"probe_score"`
	Tags           map[string]string `json:"tags,omitempty"`
}

type VideoInfo struct {
	Index          int     `json:"index"`
	Codec          string  `json:"codec"`
	CodecLongName  string  `json:"codec_long_name"`
	Profile        string  `json:"profile,omitempty"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	AspectRatio    string  `json:"aspect_ratio"`
	PixelFormat    string  `json:"pixel_format"`
	FrameRate      string  `json:"frame_rate"`
	AvgFrameRate   string  `json:"avg_frame_rate"`
	Bitrate        int64   `json:"bitrate,omitempty"`
	Duration       float64 `json:"duration,omitempty"`
	FrameCount     int64   `json:"frame_count,omitempty"`
	Level          int     `json:"level,omitempty"`
	ColorSpace     string  `json:"color_space,omitempty"`
	ColorPrimaries string  `json:"color_primaries,omitempty"`
	ColorTransfer  string  `json:"color_transfer,omitempty"`
	HasBFrames     int     `json:"has_b_frames,omitempty"`
}

type AudioInfo struct {
	Index         int     `json:"index"`
	Codec         string  `json:"codec"`
	CodecLongName string  `json:"codec_long_name"`
	Profile       string  `json:"profile,omitempty"`
	Channels      int     `json:"channels"`
	ChannelLayout string  `json:"channel_layout"`
	SampleRate    int     `json:"sample_rate"`
	SampleFormat  string  `json:"sample_format"`
	Bitrate       int64   `json:"bitrate,omitempty"`
	Duration      float64 `json:"duration,omitempty"`
}

type StreamInfo struct {
	Index     int               `json:"index"`
	Type      string            `json:"type"`
	Codec     string            `json:"codec"`
	CodecType string            `json:"codec_type"`
	Tags      map[string]string `json:"tags,omitempty"`
}

func New(options Options) *Analyzer {
	return &Analyzer{
		options: options,
		ffprobe: ffprobe.New(),
	}
}

func (a *Analyzer) Analyze(input string) (*MediaInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.options.Timeout)*time.Second)
	defer cancel()

	probeData, err := a.ffprobe.Probe(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	info := &MediaInfo{
		Input:      input,
		AnalyzedAt: time.Now(),
	}

	if a.options.ShowFormat && probeData.Format != nil {
		info.Format = a.extractFormatInfo(probeData.Format)
	}

	for _, stream := range probeData.Streams {
		switch stream.CodecType {
		case "video":
			if a.options.ShowVideo && info.VideoStream == nil {
				info.VideoStream = a.extractVideoInfo(&stream)
			}
		case "audio":
			if a.options.ShowAudio && info.AudioStream == nil {
				info.AudioStream = a.extractAudioInfo(&stream)
			}
		}

		if a.options.ShowStreams {
			info.Streams = append(info.Streams, a.extractStreamInfo(&stream))
		}
	}

	return info, nil
}

func (a *Analyzer) extractFormatInfo(format *ffprobe.Format) *FormatInfo {
	return &FormatInfo{
		FormatName:     format.FormatName,
		FormatLongName: format.FormatLongName,
		Duration:       format.Duration,
		Size:           format.SizeInt,
		Bitrate:        format.Bitrate,
		ProbeScore:     format.ProbeScore,
		Tags:           format.Tags,
	}
}

func (a *Analyzer) extractVideoInfo(stream *ffprobe.Stream) *VideoInfo {
	return &VideoInfo{
		Index:          stream.Index,
		Codec:          stream.CodecName,
		CodecLongName:  stream.CodecLongName,
		Profile:        stream.Profile,
		Width:          stream.Width,
		Height:         stream.Height,
		AspectRatio:    stream.DisplayAspectRatio,
		PixelFormat:    stream.PixFmt,
		FrameRate:      stream.RFrameRate,
		AvgFrameRate:   stream.AvgFrameRate,
		Bitrate:        stream.Bitrate,
		Duration:       stream.Duration,
		FrameCount:     stream.NbFramesInt,
		Level:          stream.Level,
		ColorSpace:     stream.ColorSpace,
		ColorPrimaries: stream.ColorPrimaries,
		ColorTransfer:  stream.ColorTransfer,
		HasBFrames:     stream.HasBFrames,
	}
}

func (a *Analyzer) extractAudioInfo(stream *ffprobe.Stream) *AudioInfo {
	sampleRate := 0
	if stream.SampleRate != "" {
		fmt.Sscanf(stream.SampleRate, "%d", &sampleRate)
	}
	
	return &AudioInfo{
		Index:         stream.Index,
		Codec:         stream.CodecName,
		CodecLongName: stream.CodecLongName,
		Profile:       stream.Profile,
		Channels:      stream.Channels,
		ChannelLayout: stream.ChannelLayout,
		SampleRate:    sampleRate,
		SampleFormat:  stream.SampleFmt,
		Bitrate:       stream.Bitrate,
		Duration:      stream.Duration,
	}
}

func (a *Analyzer) extractStreamInfo(stream *ffprobe.Stream) StreamInfo {
	return StreamInfo{
		Index:     stream.Index,
		Type:      stream.CodecType,
		Codec:     stream.CodecName,
		CodecType: stream.CodecType,
		Tags:      stream.Tags,
	}
}

// DetailedAnalysis contains extended analysis results
type DetailedAnalysis struct {
	MediaInfo       *MediaInfo              `json:"media_info"`
	Problems        []detector.Problem      `json:"problems,omitempty"`
	Packets         []PacketData            `json:"packets,omitempty"`
	Frames          []FrameData             `json:"frames,omitempty"`
	BitrateTimeline []detector.BitratePoint `json:"bitrate_timeline,omitempty"`
}

// PacketData represents analyzed packet information
type PacketData struct {
	PTS         float64 `json:"pts"`
	DTS         float64 `json:"dts"`
	Size        int     `json:"size"`
	StreamIndex int     `json:"stream_index"`
	CodecType   string  `json:"codec_type"`
	Duration    float64 `json:"duration,omitempty"`
	Flags       string  `json:"flags,omitempty"`
}

// FrameData represents analyzed frame information
type FrameData struct {
	MediaType   string  `json:"media_type"`
	StreamIndex int     `json:"stream_index"`
	KeyFrame    bool    `json:"key_frame"`
	PTS         float64 `json:"pts"`
	DTS         float64 `json:"dts"`
	Duration    float64 `json:"duration"`
	Size        int     `json:"size"`
	PictType    string  `json:"pict_type,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	PixFmt      string  `json:"pix_fmt,omitempty"`
}

// AnalyzeWithDetails performs comprehensive media analysis including packets and frames
func (a *Analyzer) AnalyzeWithDetails(input string) (*DetailedAnalysis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.options.Timeout)*time.Second)
	defer cancel()

	// Get basic media info
	mediaInfo, err := a.Analyze(input)
	if err != nil {
		return nil, err
	}

	result := &DetailedAnalysis{
		MediaInfo: mediaInfo,
		Problems:  make([]detector.Problem, 0),
	}

	// Initialize detector
	det := detector.New()

	// Analyze packets if requested
	if a.options.AnalyzePackets {
		if a.options.Verbose {
			fmt.Printf("Analyzing packets...\n")
		}
		packetsData, err := a.ffprobe.ProbePackets(ctx, input)
		if err != nil {
			if a.options.Verbose {
				fmt.Printf("Warning: Failed to analyze packets: %v\n", err)
			}
		} else {
			// Convert and limit packets
			for i, packet := range packetsData.Packets {
				if a.options.MaxPackets > 0 && i >= a.options.MaxPackets {
					break
				}
				result.Packets = append(result.Packets, PacketData{
					PTS:         packet.PTS,
					DTS:         packet.DTS,
					Size:        packet.Size,
					StreamIndex: packet.StreamIndex,
					CodecType:   packet.CodecType,
					Duration:    packet.Duration,
					Flags:       packet.Flags,
				})
			}

			// Detect packet-based problems
			if len(result.Packets) > 0 {
				packetInfos := make([]detector.PacketInfo, 0, len(result.Packets))
				for _, p := range result.Packets {
					packetInfos = append(packetInfos, detector.PacketInfo{
						PTS:         p.PTS,
						DTS:         p.DTS,
						Size:        p.Size,
						StreamIndex: p.StreamIndex,
						Duration:    p.Duration,
					})
				}
				det.DetectBitrateVariations(packetInfos)
				det.DetectPacketLoss(packetInfos)
				
				// Generate bitrate timeline
				result.BitrateTimeline = detector.GenerateBitrateTimeline(packetInfos, 1.0)
			}
		}
	}

	// Analyze frames if requested
	if a.options.AnalyzeFrames {
		if a.options.Verbose {
			fmt.Printf("Analyzing frames...\n")
		}
		framesData, err := a.ffprobe.ProbeFrames(ctx, input)
		if err != nil {
			if a.options.Verbose {
				fmt.Printf("Warning: Failed to analyze frames: %v\n", err)
			}
		} else {
			// Convert and limit frames
			for i, frame := range framesData.Frames {
				if a.options.MaxFrames > 0 && i >= a.options.MaxFrames {
					break
				}
				result.Frames = append(result.Frames, FrameData{
					MediaType:   frame.MediaType,
					StreamIndex: frame.StreamIndex,
					KeyFrame:    frame.KeyFrame,
					PTS:         frame.PTS,
					DTS:         frame.DTS,
					Duration:    frame.Duration,
					Size:        frame.Size,
					PictType:    frame.PictType,
					Width:       frame.Width,
					Height:      frame.Height,
					PixFmt:      frame.PixFmt,
				})
			}

			// Detect frame-based problems
			if len(result.Frames) > 0 {
				frameInfos := make([]detector.FrameInfo, 0, len(result.Frames))
				for _, f := range result.Frames {
					frameInfos = append(frameInfos, detector.FrameInfo{
						MediaType:   f.MediaType,
						StreamIndex: f.StreamIndex,
						KeyFrame:    f.KeyFrame,
						PTS:         f.PTS,
						DTS:         f.DTS,
						Duration:    f.Duration,
						Size:        f.Size,
						PictType:    f.PictType,
					})
				}
				det.DetectKeyframeIssues(frameInfos)
				det.DetectTimestampIssues(frameInfos)
			}
		}
	}

	// Check compatibility issues
	if mediaInfo.VideoStream != nil {
		det.AnalyzeCompatibility(
			mediaInfo.VideoStream.Codec,
			mediaInfo.VideoStream.Profile,
			mediaInfo.VideoStream.Level,
			mediaInfo.Format.FormatName,
		)
	}

	result.Problems = det.GetProblems()

	return result, nil
}