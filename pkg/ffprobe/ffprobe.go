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

// ProbePackets extracts packet information from media file
func (f *FFProbe) ProbePackets(ctx context.Context, input string) (*PacketsData, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_packets",
		input,
	}

	cmd := exec.CommandContext(ctx, f.binary, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("ffprobe packets failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run ffprobe for packets: %w", err)
	}

	var data PacketsData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse packets output: %w", err)
	}

	// Convert string timestamps to float64
	for i := range data.Packets {
		packet := &data.Packets[i]
		if packet.PTSTime != "" {
			if pts, err := strconv.ParseFloat(packet.PTSTime, 64); err == nil {
				packet.PTS = pts
			}
		}
		if packet.DTSTime != "" {
			if dts, err := strconv.ParseFloat(packet.DTSTime, 64); err == nil {
				packet.DTS = dts
			}
		}
		if packet.DurationTime != "" {
			if duration, err := strconv.ParseFloat(packet.DurationTime, 64); err == nil {
				packet.Duration = duration
			}
		}
		if packet.SizeStr != "" {
			if size, err := strconv.Atoi(packet.SizeStr); err == nil {
				packet.Size = size
			}
		}
	}

	return &data, nil
}

// ProbeFrames extracts frame information from media file
func (f *FFProbe) ProbeFrames(ctx context.Context, input string) (*FramesData, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_frames",
		input,
	}

	cmd := exec.CommandContext(ctx, f.binary, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("ffprobe frames failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run ffprobe for frames: %w", err)
	}

	var data FramesData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse frames output: %w", err)
	}

	// Convert string values to appropriate types
	for i := range data.Frames {
		frame := &data.Frames[i]
		if frame.PTSTime != "" {
			if pts, err := strconv.ParseFloat(frame.PTSTime, 64); err == nil {
				frame.PTS = pts
			}
		}
		if frame.DTSTime != "" {
			if dts, err := strconv.ParseFloat(frame.DTSTime, 64); err == nil {
				frame.DTS = dts
			}
		}
		if frame.DurationTime != "" {
			if duration, err := strconv.ParseFloat(frame.DurationTime, 64); err == nil {
				frame.Duration = duration
			}
		}
		if frame.PktSize != "" {
			if size, err := strconv.Atoi(frame.PktSize); err == nil {
				frame.Size = size
			}
		}
		frame.KeyFrame = frame.KeyFrameInt == 1
	}

	return &data, nil
}

// PacketsData holds packet information
type PacketsData struct {
	Packets []Packet `json:"packets"`
}

// Packet represents a media packet
type Packet struct {
	CodecType    string  `json:"codec_type"`
	StreamIndex  int     `json:"stream_index"`
	PTS          float64 `json:"-"`
	PTSTime      string  `json:"pts_time"`
	DTS          float64 `json:"-"`
	DTSTime      string  `json:"dts_time"`
	Duration     float64 `json:"-"`
	DurationTime string  `json:"duration_time"`
	Size         int     `json:"-"`
	SizeStr      string  `json:"size"`
	Pos          string  `json:"pos"`
	Flags        string  `json:"flags"`
}

// FramesData holds frame information
type FramesData struct {
	Frames []Frame `json:"frames"`
}

// Frame represents a media frame
type Frame struct {
	MediaType           string  `json:"media_type"`
	StreamIndex         int     `json:"stream_index"`
	KeyFrameInt         int     `json:"key_frame"`
	KeyFrame            bool    `json:"-"`
	PTS                 float64 `json:"-"`
	PTSTime             string  `json:"pts_time"`
	DTS                 float64 `json:"-"`
	DTSTime             string  `json:"dts_time"`
	Duration            float64 `json:"-"`
	DurationTime        string  `json:"duration_time"`
	Size                int     `json:"-"`
	PktSize             string  `json:"pkt_size"`
	Width               int     `json:"width,omitempty"`
	Height              int     `json:"height,omitempty"`
	PixFmt              string  `json:"pix_fmt,omitempty"`
	PictType            string  `json:"pict_type,omitempty"`
	CodedPictureNumber  int     `json:"coded_picture_number,omitempty"`
	DisplayPictureNumber int    `json:"display_picture_number,omitempty"`
	InterlacedFrame     int     `json:"interlaced_frame,omitempty"`
	TopFieldFirst       int     `json:"top_field_first,omitempty"`
	RepeatPict          int     `json:"repeat_pict,omitempty"`
}

func (f *FFProbe) CheckInstalled() error {
	cmd := exec.Command(f.binary, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found. Please install FFmpeg: %w", err)
	}
	return nil
}