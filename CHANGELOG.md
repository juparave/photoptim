# Changelog

All notable changes to Photoptim will be documented in this file.

## [Unreleased]

### Added
- SFTP host key verification against `~/.config/photoptim/known_hosts`
  (trust-on-first-use; rejects changed keys).
- Atomic remote writes: optimized files are written to a temp file and renamed
  over the original, so a failed transfer leaves the original intact.
- SFTP browser: adjustable JPEG quality (`+`/`-`), select-all (`a`) and clear
  (`c`), and large-file (≥10 MB) highlighting.
- Local TUI: select-all (`a`) / clear (`c`), per-file progress, rolling result
  list, and explicit reporting of unsupported files.

### Fixed
- SFTP error screen no longer swallows the keystroke used to dismiss it; Ctrl+C
  always quits.
- Local TUI progress bar now advances per file instead of jumping straight to
  done.
- Surface the final write/rename error from the SFTP and pipeline upload paths
  instead of reporting false success.

### Docs
- Rewrote `TUI_USAGE.md` to match the implemented behavior (the previous version
  documented many unimplemented features). Corrected the SFTP launch command in
  `README.md` (`photoptim sftp`, not `sftp-tui`).

## [v0.1.1] - 2025-08-27

### Fixed
- Corrected `go install` instructions in `README.md`.
- Fixed module path in `go.mod` to align with the GitHub repository path.

## [v0.1.0] - 2025-08-27

### Added
- Image resizing functionality with CatmullRom interpolation
- Mobile device size presets (iPhone, Samsung Galaxy, Google Pixel, iPad models)
- SFTP TUI resize parameter toggle with 'r' key
- Aspect ratio preserving resizing algorithm

### Enhanced
- SFTP TUI now displays current resize settings in status bar
- Progress tracking for batch optimization operations
- Improved error handling for unsupported file formats

### Fixed
- JPEG format detection issue in SFTP TUI
- File extension parsing consistency across optimization functions

## Future Work

### Integration Needed
- **TUI Integration**: Merge the local file TUI (`cmd/tui/main.go`) and SFTP TUI (`cmd/photoptim/main.go sftp-tui`) into a unified interface
- **Common Codebase**: Share optimization logic, UI components, and state management between both TUI applications
- **Unified Navigation**: Single application with mode switching between local and remote file operations
- **Shared Configuration**: Common settings and preferences across both interfaces

### Planned Features
- WebP format support
- More resize presets and custom dimension input
- Background optimization queue
- Performance metrics and optimization history
- Plugin system for additional image processors