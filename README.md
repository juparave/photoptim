# Photoptim

Photoptim is a fast and flexible tool for optimizing images. It supports various formats and optimization techniques.

## Features

- Optimize JPEG and PNG images
- Adjustable quality settings for JPEG compression
- Batch processing of multiple images
- CLI interface for easy usage
- TUI (Terminal User Interface) for interactive usage

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

Photoptim also includes a TUI (Terminal User Interface) for interactive usage:

```bash
go run cmd/tui/main.go
```

Or build and run the TUI application:

```bash
go build -o photoptim-tui cmd/tui/main.go
./photoptim-tui
```

See [TUI Usage Guide](TUI_USAGE.md) for detailed instructions on how to use the TUI.

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