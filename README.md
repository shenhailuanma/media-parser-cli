# media-parser-cli

A command-line tool for analyzing video files and streams, providing detailed media information for debugging and understanding media properties.

## Features

- **Comprehensive Media Analysis**: Extract detailed information about video and audio streams
- **Problem Detection**: Automatically detect common media issues and quality problems
- **Packet/Frame Analysis**: Deep dive into packet and frame-level information
- **Bitrate Analysis**: Visualize bitrate variations and detect spikes
- **Frame Type Visualization**: Eyecard-style frame type analysis (I/P/B frames)
- **Export Capabilities**: Save detailed analysis results to JSON files for further processing
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
# Analyze a local video file with problem detection
media-parser-cli parse video.mp4

# Analyze an HLS stream
media-parser-cli parse https://example.com/stream.m3u8

# Analyze an RTMP stream
media-parser-cli parse rtmp://server/live/stream

# Export detailed analysis to JSON files
media-parser-cli export video.mp4 -d ./analysis_output
```

### Commands

#### parse - Quick Media Analysis
```bash
media-parser-cli parse [options] <input>

Options:
  --show-video        Show video stream information (default: true)
  --show-audio        Show audio stream information (default: true)
  --show-format       Show container format information (default: true)
  --show-streams      Show all stream details (default: false)
  --show-problems     Show detected problems and warnings (default: true)
  --show-all          Show all available information
  -o, --output        Output format: json, yaml, text (default: text)
  -v, --verbose       Enable verbose output
  --timeout           Analysis timeout in seconds (default: 30)
  -h, --help          Show help information
```

#### export - Detailed Analysis Export
```bash
media-parser-cli export [options] <input>

Options:
  -d, --dir           Directory to save analysis files (default: ./media-analysis)
  --export-packets    Export packet information
  --export-frames     Export frame information
  --export-problems   Export detected problems (default: true)
  --export-bitrate    Export bitrate timeline
  --export-all        Export all available information
  --max-packets       Maximum number of packets to export (default: 10000)
  --max-frames        Maximum number of frames to export (default: 5000)
  -v, --verbose       Enable verbose output
  --timeout           Analysis timeout in seconds (default: 30)
```

### Examples

#### Basic analysis with problem detection
```bash
media-parser-cli parse video.mp4
```

#### Get JSON output for automation
```bash
media-parser-cli parse video.mp4 -o json
```

#### Export complete analysis
```bash
media-parser-cli export video.mp4 -d ./reports --export-all
```

#### Export frame analysis (like eyecard)
```bash
media-parser-cli export video.mp4 -d ./debug --export-frames --max-frames 1000
```

#### Disable problem detection for faster analysis
```bash
media-parser-cli parse video.mp4 --show-problems=false
```

#### Verbose mode with all information
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

DETECTED PROBLEMS:
----------------------------------------

🟡 WARNINGS:
  [BITRATE_HIGH_VARIANCE]  High bitrate variation detected (CV: 0.35)
    Details:              Average: 1.95 Mbps, StdDev: 0.68 Mbps
    ✨ Suggestion:        Consider using constant bitrate encoding or adjusting rate control settings

  [LARGE_KEYFRAME_INTERVAL]  Large keyframe interval: 10.50s
    ✨ Suggestion:        For streaming, consider reducing keyframe interval to 2-4 seconds

----------------------------------------
Summary: 0 errors, 0 critical, 2 warnings, 0 info
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
│   ├── parse.go           # Parse command implementation
│   └── export.go          # Export command for detailed analysis
├── internal/
│   ├── analyzer/          # Media analysis logic
│   ├── detector/          # Problem detection engine
│   └── reporter/          # Output formatting
├── pkg/
│   └── ffprobe/          # FFprobe wrapper with packet/frame analysis
└── main.go               # Application entry point
```

### Problem Detection

The tool automatically detects various media issues:

- **Bitrate Issues**: High variance, sudden spikes
- **Keyframe Problems**: Irregular intervals, missing keyframes
- **Timestamp Issues**: Non-monotonic PTS/DTS, large gaps
- **Compatibility Issues**: Codec/container compatibility warnings
- **Packet Loss Indicators**: Potential packet loss detection

### Export Files

The export command creates structured JSON files:

- `media_info.json`: Basic media information
- `problems.json`: Detected issues with severity levels
- `packets.json`: Packet-level data for detailed analysis
- `frames.json`: Frame-level information
- `frame_visualization.json`: Eyecard-style frame type visualization
- `bitrate_timeline.json`: Bitrate over time for visualization
- `summary.json`: Export summary and statistics

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
