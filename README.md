# Photoptim

Photoptim is a fast and flexible tool for optimizing images. It supports various formats and optimization techniques.

## Features

- Optimize JPEG and PNG images
- Adjustable quality settings for JPEG compression
- Image resizing with aspect ratio preservation
- Multiple mobile device size presets (iPhone, Samsung, iPad, etc.)
- Batch processing of multiple images
- SFTP remote file optimization
- CLI interface for easy usage
- TUI (Terminal User Interface) for interactive usage
- SFTP TUI for remote file management and optimization

## Installation

To install Photoptim, you need to have Go installed on your system. Then run:

```bash
go install github.com/yourusername/photoptim@latest
```

Or clone this repository and build it locally:

```bash
git clone https://github.com/yourusername/photoptim.git
cd photoptim
go build -o photoptim cmd/photoptim/main.go
```

## Usage

### Command Line Interface

#### Optimize a single image

```bash
photoptim optimize input.jpg output.jpg --quality 80
```

#### Batch optimize images

```bash
photoptim batch input_directory output_directory --quality 75
```

### Terminal User Interface

Photoptim includes two TUI (Terminal User Interface) applications:

#### Local File TUI
For optimizing local images:

```bash
go run cmd/tui/main.go
```

Or build and run the TUI application:

```bash
go build -o photoptim-tui cmd/tui/main.go
./photoptim-tui
```

#### SFTP TUI  
For optimizing remote images over SFTP:

```bash
go run cmd/photoptim/main.go sftp-tui
```

Features:
- Browse remote directories
- Select multiple files for optimization  
- Resize images to common mobile device sizes
- Real-time optimization results

See [TUI Usage Guide](TUI_USAGE.md) and [SFTP Extension Guide](SFTP_EXTENSION_PRD.md) for detailed instructions.

Note: The TUI application requires a proper terminal environment with /dev/tty access. It may not work in some IDEs or remote environments that don't provide full TTY support.

### Help

For more information about commands and options:

```bash
photoptim --help
photoptim optimize --help
photoptim batch --help
```

## License

This project is licensed under the MIT License.