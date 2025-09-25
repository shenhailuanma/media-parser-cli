# media-parser-cli

A command-line tool for analyzing video files and streams, providing detailed media information for debugging and understanding media properties.

## Features

- **Comprehensive Media Analysis**: Extract detailed information about video and audio streams
- **Multiple Input Support**: Analyze local files, HTTP/HTTPS streams, HLS, DASH, RTMP, and RTSP
- **Flexible Output Formats**: JSON, YAML, or human-readable text reports
- **Stream Information**: Codec details, resolution, bitrate, frame rate, and more
- **Container Format Details**: Duration, file size, overall bitrate
- **Fast Analysis**: Configurable timeout for quick results

## Installation

### Prerequisites

- Go 1.19 or higher
- FFmpeg/ffprobe installed on your system

#### Install FFmpeg

**macOS:**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows:**
Download from [FFmpeg official website](https://ffmpeg.org/download.html)

### Build from Source

```bash
git clone https://github.com/tomi/media-parser-cli.git
cd media-parser-cli
go build -o media-parser-cli
```

## Usage

### Basic Usage

```bash
# Analyze a local video file
media-parser-cli parse video.mp4

# Analyze an HLS stream
media-parser-cli parse https://example.com/stream.m3u8

# Analyze an RTMP stream
media-parser-cli parse rtmp://server/live/stream
```

### Command Options

```bash
media-parser-cli parse [options] <input>

Options:
  --show-video        Show video stream information (default: true)
  --show-audio        Show audio stream information (default: true)
  --show-format       Show container format information (default: true)
  --show-streams      Show all stream details (default: false)
  --show-all          Show all available information
  -o, --output        Output format: json, yaml, text (default: text)
  -v, --verbose       Enable verbose output
  --timeout           Analysis timeout in seconds (default: 30)
  -h, --help          Show help information
```

### Examples

#### Get JSON output for automation
```bash
media-parser-cli parse video.mp4 -o json
```

#### Show all stream information
```bash
media-parser-cli parse video.mp4 --show-all
```

#### Quick analysis with 10-second timeout
```bash
media-parser-cli parse https://stream.example.com/live.m3u8 --timeout 10
```

#### Verbose mode for debugging
```bash
media-parser-cli parse video.mp4 -v --show-all
```

## Output Examples

### Text Output (Default)
```
================================================================================
MEDIA ANALYSIS REPORT
Analyzed at: 2024-01-25T10:30:00Z
Input: sample.mp4
================================================================================

CONTAINER FORMAT:
----------------------------------------
Format:           mp4
Long Name:        MP4 (MPEG-4 Part 14)
Duration:         00:10:35.640
File Size:        156.78 MB
Overall Bitrate:  2.07 Mbps

VIDEO STREAM:
----------------------------------------
Stream Index:     0
Codec:            h264 (H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10)
Profile:          High
Resolution:       1920x1080
Frame Rate:       29.97 fps
Bitrate:          1.95 Mbps

AUDIO STREAM:
----------------------------------------
Stream Index:     1
Codec:            aac (AAC (Advanced Audio Coding))
Channels:         2
Sample Rate:      48000 Hz
Bitrate:          128.0 kbps
```

### JSON Output
```json
{
  "input": "sample.mp4",
  "format": {
    "format_name": "mp4",
    "format_long_name": "MP4 (MPEG-4 Part 14)",
    "duration": 635.64,
    "size": 164405248,
    "bitrate": 2069035
  },
  "video": {
    "index": 0,
    "codec": "h264",
    "width": 1920,
    "height": 1080,
    "frame_rate": "30000/1001",
    "bitrate": 1950000
  }
}
```

## Development

### Project Structure
```
media-parser-cli/
├── cmd/                    # Command definitions
│   ├── root.go            # Root command setup
│   └── parse.go           # Parse command implementation
├── internal/
│   ├── analyzer/          # Media analysis logic
│   └── reporter/          # Output formatting
├── pkg/
│   └── ffprobe/          # FFprobe wrapper
└── main.go               # Application entry point
```

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o media-parser-cli
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
