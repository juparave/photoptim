<div align="center">

<!-- Placeholder for Logo -->
<!-- <img src="docs/logo.png" alt="Photoptim Logo" width="200"/> -->

# Photoptim

**A fast, flexible, and interactive tool for optimizing images locally and remotely.**

[![Go Version](https://img.shields.io/github/go-mod/go-version/juparave/photoptim)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/juparave/photoptim)](https://goreportcard.com/report/github.com/juparave/photoptim)
<!-- Uncomment when CI is set up -->
<!-- [![Build Status](https://img.shields.io/github/actions/workflow/status/juparave/photoptim/ci.yml?branch=main)](https://github.com/juparave/photoptim/actions) -->

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#tech-stack">Tech Stack</a>
</p>

</div>

---

Photoptim is a powerful utility designed to reduce image file sizes without significant quality loss. It provides both a traditional Command Line Interface (CLI) for automation and scripts, and a rich Terminal User Interface (TUI) for interactive use—including remote file management via SFTP.

## ✨ Features

- **Format Support:** Optimize JPEG and PNG images efficiently.
- **Adjustable Compression:** Fine-tune quality settings for JPEG compression.
- **Smart Resizing:** Resize images while preserving aspect ratios.
- **Device Presets:** Built-in mobile device size presets (iPhone, Samsung, iPad, etc.).
- **Batch Processing:** Easily process multiple images in bulk.
- **SFTP Remote Optimization:** Manage and optimize files directly on remote servers via SFTP.
- **Dual Interfaces:** 
  - Standard **CLI** for quick commands and automation.
  - Interactive **TUI** for browsing, selecting, and optimizing files locally or remotely.

---

## 📸 Screenshots

<!-- Add your screenshots here -->
> **Note:** Screenshots coming soon!
> 
> *CLI Example:*
> `photoptim batch input/ output/ --quality 80`
> 
> *TUI Example:*
> Browse and select files interactively with `photoptim tui`.

---

## 🚀 Installation

Ensure you have Go installed on your system. This project provides two commands: the main `photoptim` CLI and the interactive `photoptim-tui`.

### Using `go install` (Recommended)

```bash
# Install the main CLI tool
go install github.com/juparave/photoptim/cmd/photoptim@latest

# Install the TUI (Terminal User Interface)
go install github.com/juparave/photoptim/cmd/tui@latest
```
*Note: The TUI will be installed as `tui`. You may want to rename it to `photoptim-tui` or create an alias for clarity.*

### Building from Source

```bash
git clone https://github.com/juparave/photoptim.git
cd photoptim

# Build the CLI
go build -o photoptim cmd/photoptim/main.go

# Build the TUI
go build -o photoptim-tui cmd/tui/main.go
```

---

## 🛠 Usage

### Command Line Interface (CLI)

**Optimize a single image:**
```bash
photoptim optimize input.jpg output.jpg --quality 80
```

**Batch optimize images:**
```bash
photoptim batch ./input_dir ./output_dir --quality 75
```

### Terminal User Interface (TUI)

Photoptim features two distinct TUI applications:

**1. Local File TUI:**
For optimizing images on your local machine interactively.
```bash
go run cmd/tui/main.go
# Or if built:
./photoptim-tui
```

**2. SFTP TUI:**
For browsing and optimizing remote images over SFTP.
```bash
go run cmd/photoptim/main.go sftp-tui
# Or if built:
./photoptim sftp-tui
```
*Features include: Remote directory browsing, multi-select optimization, and real-time result feedback.*

For detailed usage, check out the [TUI Usage Guide](TUI_USAGE.md) and [SFTP Extension Guide](SFTP_EXTENSION_PRD.md).

> ⚠️ **Note:** The TUI requires a proper terminal environment with `/dev/tty` access. It may not function correctly in some IDEs or limited remote environments.

### Help

```bash
photoptim --help
photoptim optimize --help
photoptim batch --help
```

---

## ⚙️ Tech Stack

Photoptim is built using modern Go libraries to ensure performance and a great developer/user experience:

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework.
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) & [Lip Gloss](https://github.com/charmbracelet/lipgloss) - For building the beautiful and robust TUI.
- [sftp](https://github.com/pkg/sftp) - For secure remote file operations.
- [bbolt](https://github.com/etcd-io/bbolt) - For fast, reliable local caching.
- [golang.org/x/image](https://golang.org/x/image) - For advanced image processing.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
