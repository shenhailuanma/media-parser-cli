# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Building the project
```bash
go build -o media-parser-cli
```

### Running tests
```bash
go test ./...
go test -v ./...  # verbose output
go test -cover ./...  # with coverage
```

### Installing dependencies
```bash
go mod download
go mod tidy  # clean up and verify dependencies
```

### Running the application
```bash
# After building
./media-parser-cli parse <video-file-or-stream>

# Without building
go run main.go parse <video-file-or-stream>
```

### Development workflow
```bash
# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Check for vulnerabilities
go list -json -deps ./... | nancy sleuth
```

## Project Architecture

### Core Components

1. **CLI Framework (Cobra)**
   - Entry point: `main.go`
   - Command definitions: `cmd/` directory
   - Root command sets up global flags and configuration
   - Parse command handles media analysis requests

2. **Media Analyzer (`internal/analyzer/`)**
   - Orchestrates the media analysis process
   - Interfaces with FFprobe for media extraction
   - Transforms raw probe data into structured MediaInfo
   - Handles timeouts and error conditions

3. **FFprobe Wrapper (`pkg/ffprobe/`)**
   - Low-level interface to FFmpeg's ffprobe tool
   - Executes ffprobe with JSON output format
   - Parses and structures the probe results
   - Handles binary detection and error reporting

4. **Reporter (`internal/reporter/`)**
   - Formats analysis results for output
   - Supports multiple output formats (text, JSON, YAML)
   - Handles human-readable formatting for text output
   - Provides consistent output structure across formats

### Data Flow

1. User invokes `media-parser-cli parse <input>`
2. Cobra parses command-line arguments and flags
3. Parse command creates an Analyzer with specified options
4. Analyzer uses FFprobe wrapper to extract media information
5. Raw FFprobe data is transformed into structured MediaInfo
6. Reporter formats MediaInfo based on requested output format
7. Formatted result is written to stdout

### Key Design Decisions

- **FFprobe Dependency**: Uses FFmpeg's ffprobe as the underlying analysis engine for broad format support and reliability
- **Modular Architecture**: Clear separation between CLI, analysis, and reporting concerns for maintainability
- **Context-based Timeouts**: Uses Go's context package for proper timeout handling during analysis
- **Structured Output**: Consistent data structures enable easy consumption by other tools when using JSON/YAML output

## Testing Strategy

- Unit tests for individual components (analyzer, reporter, ffprobe wrapper)
- Integration tests with sample media files
- Mock FFprobe responses for testing error conditions
- Benchmark tests for performance-critical paths

## External Dependencies

- **github.com/spf13/cobra**: CLI framework for command parsing and help generation
- **FFmpeg/ffprobe**: Required system dependency for media analysis