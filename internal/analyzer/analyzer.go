package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/tomi/media-parser-cli/pkg/ffprobe"
)

type Options struct {
	Timeout     int
	ShowVideo   bool
	ShowAudio   bool
	ShowFormat  bool
	ShowStreams bool
	Verbose     bool
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