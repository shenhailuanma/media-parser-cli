package ffprobe

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type FFProbe struct {
	binary string
}

type ProbeData struct {
	Streams []Stream `json:"streams"`
	Format  *Format  `json:"format"`
}

type Stream struct {
	Index              int               `json:"index"`
	CodecName          string            `json:"codec_name"`
	CodecLongName      string            `json:"codec_long_name"`
	Profile            string            `json:"profile"`
	CodecType          string            `json:"codec_type"`
	CodecTimeBase      string            `json:"codec_time_base"`
	CodecTagString     string            `json:"codec_tag_string"`
	CodecTag           string            `json:"codec_tag"`
	Width              int               `json:"width,omitempty"`
	Height             int               `json:"height,omitempty"`
	CodedWidth         int               `json:"coded_width,omitempty"`
	CodedHeight        int               `json:"coded_height,omitempty"`
	HasBFrames         int               `json:"has_b_frames,omitempty"`
	PixFmt             string            `json:"pix_fmt,omitempty"`
	Level              int               `json:"level,omitempty"`
	ColorRange         string            `json:"color_range,omitempty"`
	ColorSpace         string            `json:"color_space,omitempty"`
	ColorTransfer      string            `json:"color_transfer,omitempty"`
	ColorPrimaries     string            `json:"color_primaries,omitempty"`
	ChromaLocation     string            `json:"chroma_location,omitempty"`
	Refs               int               `json:"refs,omitempty"`
	IsAVC              string            `json:"is_avc,omitempty"`
	NalLengthSize      string            `json:"nal_length_size,omitempty"`
	RFrameRate         string            `json:"r_frame_rate,omitempty"`
	AvgFrameRate       string            `json:"avg_frame_rate,omitempty"`
	TimeBase           string            `json:"time_base"`
	StartPts           int64             `json:"start_pts"`
	StartTime          string            `json:"start_time"`
	DurationTs         int64             `json:"duration_ts"`
	Duration           float64           `json:"duration,string"`
	BitRate            string            `json:"bit_rate,omitempty"`
	BitsPerRawSample   string            `json:"bits_per_raw_sample,omitempty"`
	NbFrames           string            `json:"nb_frames,omitempty"`
	DisplayAspectRatio string            `json:"display_aspect_ratio,omitempty"`
	SampleFmt          string            `json:"sample_fmt,omitempty"`
	SampleRate         string            `json:"sample_rate,omitempty"`
	Channels           int               `json:"channels,omitempty"`
	ChannelLayout      string            `json:"channel_layout,omitempty"`
	BitsPerSample      int               `json:"bits_per_sample,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
	Bitrate            int64
	NbFramesInt        int64
}

type Format struct {
	Filename       string            `json:"filename"`
	NbStreams      int               `json:"nb_streams"`
	NbPrograms     int               `json:"nb_programs"`
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	StartTime      string            `json:"start_time"`
	Duration       float64           `json:"duration,string"`
	Size           string            `json:"size"`
	BitRate        string            `json:"bit_rate"`
	ProbeScore     int               `json:"probe_score"`
	Tags           map[string]string `json:"tags,omitempty"`
	Bitrate        int64
	SizeInt        int64
}

func New() *FFProbe {
	return &FFProbe{
		binary: "ffprobe",
	}
}

func (f *FFProbe) Probe(ctx context.Context, input string) (*ProbeData, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		input,
	}

	cmd := exec.CommandContext(ctx, f.binary, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("ffprobe failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var data ProbeData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	for i := range data.Streams {
		stream := &data.Streams[i]
		if stream.BitRate != "" {
			if bitrate, err := strconv.ParseInt(stream.BitRate, 10, 64); err == nil {
				stream.Bitrate = bitrate
			}
		}
		if stream.NbFrames != "" {
			if frames, err := strconv.ParseInt(stream.NbFrames, 10, 64); err == nil {
				stream.NbFramesInt = frames
			}
		}
		if stream.SampleRate != "" {
			if sampleRate, err := strconv.Atoi(stream.SampleRate); err == nil {
				stream.SampleRate = strconv.Itoa(sampleRate)
			}
		}
	}

	if data.Format != nil {
		if data.Format.BitRate != "" {
			if bitrate, err := strconv.ParseInt(data.Format.BitRate, 10, 64); err == nil {
				data.Format.Bitrate = bitrate
			}
		}
		if data.Format.Size != "" {
			if size, err := strconv.ParseInt(data.Format.Size, 10, 64); err == nil {
				data.Format.SizeInt = size
			}
		}
	}

	return &data, nil
}

func (f *FFProbe) CheckInstalled() error {
	cmd := exec.Command(f.binary, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found. Please install FFmpeg: %w", err)
	}
	return nil
}