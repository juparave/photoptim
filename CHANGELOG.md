# Changelog

All notable changes to Photoptim will be documented in this file.

## [Unreleased]

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