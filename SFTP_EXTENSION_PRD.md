# Photoptim SFTP Extension - Product Requirements Document

## Overview

### Purpose
This document outlines the requirements for extending the Photoptim application with SFTP capabilities, enabling users to remotely browse, select, optimize, and upload images directly from a remote media server through a Bubble Tea-based TUI.

### Scope
The SFTP extension will add a new workflow to Photoptim that allows users to:
1. Connect to a remote SFTP server using SSH key authentication
2. Browse remote directories in a TUI file explorer
3. Filter and display images based on file size criteria
4. Select images for optimization
5. Optimize selected images locally
6. Upload optimized images back to the remote server

## User Experience

### Target Users
- Photographers managing large remote media libraries
- System administrators handling media optimization tasks
- Content creators working with remote storage solutions

### User Stories

1. As a photographer, I want to connect to my remote media server so I can optimize images without downloading them first.
2. As a user, I want to browse remote directories in a familiar file explorer interface so I can easily find my images.
3. As a content creator, I want to filter images by size so I can focus on optimizing only large files that need compression.
4. As a system administrator, I want to select multiple images for batch optimization so I can process many files efficiently.
5. As a user, I want to see a progress indicator during optimization and upload processes so I know the application is working.

## Technical Requirements

### SFTP Integration

#### Connection Management
- **Authentication**: Use local SSH keys (`~/.ssh/id_rsa`) for authentication
- **Connection Configuration**: Allow users to specify hostname, port, username, and remote path
- **Connection Status**: Display connection status and handle disconnections gracefully

#### Remote File Operations
- **Directory Browsing**: List directories and files with metadata (name, size, type)
- **File Filtering**: Filter files by extension (.jpg, .jpeg, .png) and size threshold
- **File Selection**: Allow multiple file selection with standard keyboard shortcuts
- **File Download**: Download selected files to a temporary local directory for optimization
- **File Upload**: Upload optimized files back to the remote server, optionally preserving original files

### TUI Design with Bubble Tea

#### UI Components

1. **Connection Screen**
   - Hostname input field
   - Port input field (default: 22)
   - Username input field
   - Remote path input field (default: /)
   - Connect button
   - Status messages for connection attempts

2. **File Browser Screen**
   - Directory tree navigation
   - File listing with columns:
     - Filename
     - File size (human-readable format)
     - File type (directory/file)
     - Modification date
   - Size filter input field
   - Refresh button
   - Navigation controls (up/down, enter to open, backspace to go up)

3. **File Selection Screen**
   - Multi-select file listing
   - Toggle selection with spacebar
   - Select all/none options
   - Clear selection button
   - Selected files counter
   - Proceed to optimization button

4. **Optimization Configuration Screen**
   - Quality setting slider/input (1-100)
   - Optimization mode selection (single/batch)
   - Output directory specification
   - Preserve originals toggle
   - Start optimization button

5. **Progress Screen**
   - Progress bar for overall process
   - Individual file progress indicators
   - Status messages for each operation
   - Cancel button (with confirmation)
   - Detailed log view (toggleable)

6. **Results Screen**
   - Summary of optimized files
   - Space savings statistics
   - Upload status for each file
   - Retry failed uploads option
   - Return to file browser button

#### Navigation Flow
```
[Connection Screen]
       ↓
[File Browser Screen]
       ↓
[File Selection Screen]
       ↓
[Optimization Configuration Screen]
       ↓
[Progress Screen]
       ↓
[Results Screen]
```

#### Keyboard Shortcuts
- **Global**: Ctrl+C to quit, Ctrl+L to refresh
- **Navigation**: Arrow keys for selection, Enter to confirm, Esc to go back
- **File Browser**: Space to select/deselect, A to select all, N to select none
- **Progress**: Ctrl+X to cancel (with confirmation)

### Bubble Tea Implementation Details

#### State Management
The application will use a state machine pattern with the following states:
1. `ConnectionState` - Managing SFTP connection parameters
2. `BrowserState` - Browsing remote directories
3. `SelectionState` - Selecting files for optimization
4. `ConfigState` - Configuring optimization parameters
5. `ProgressState` - Displaying optimization and upload progress
6. `ResultsState` - Showing final results

#### Model Structure
```go
type Model struct {
    state          State
    sftpClient     *sftp.Client
    connection     ConnectionConfig
    currentPath    string
    files          []RemoteFile
    selectedFiles  []RemoteFile
    filterSize     int64
    quality        int
    progress       progress.Model
    logs           []string
    width, height  int
}
```

#### Key Components
1. **File Picker Component**: Custom implementation for remote files
2. **Progress Bar**: Using `bubbles/progress` for optimization and upload progress
3. **Text Inputs**: Using `bubbles/textinput` for configuration fields
4. **Lists**: Using `bubbles/list` for file listings and selections
5. **Status Messages**: Custom implementation for operation feedback

#### Asynchronous Operations
- **SFTP Operations**: File listing, download, and upload operations will be implemented as Bubble Tea commands
- **Optimization**: Image optimization will run as background commands
- **Progress Updates**: Real-time progress updates through custom messages

### Performance Considerations

#### Caching
- Cache directory listings to reduce SFTP requests
- Cache file metadata to improve browsing performance
- Implement cache expiration based on modification times

#### Memory Management
- Stream large files during download/upload to minimize memory usage
- Process images in batches to prevent memory exhaustion
- Clean up temporary files after successful operations

#### Error Handling
- Graceful handling of network interruptions
- Retry mechanisms for failed operations
- Detailed error messages for troubleshooting
- Recovery options for partial operations

### Security Requirements

#### Authentication
- Use SSH key authentication only (no password support)
- Validate SSH key permissions and format
- Support for standard SSH key locations
- Secure handling of connection credentials

#### Data Protection
- Encrypt temporary files during processing
- Secure deletion of temporary files
- Validate file paths to prevent directory traversal
- Limit file operations to specified directories

## Implementation Phases

### Phase 1: Core SFTP Integration
- Implement SFTP connection management
- Create basic file listing functionality
- Add file download/upload capabilities

### Phase 2: TUI Implementation
- Design and implement Bubble Tea models for each screen
- Create navigation between screens
- Add basic keyboard controls

### Phase 3: Advanced Features
- Implement file filtering by size
- Add multi-file selection capabilities
- Create progress tracking and status updates

### Phase 4: Polish and Optimization
- Add caching for improved performance
- Implement comprehensive error handling
- Add detailed logging and debugging features
- Optimize memory usage for large file operations

## Testing Requirements

### Unit Tests
- SFTP connection and authentication
- File operations (list, download, upload)
- File filtering logic
- Progress tracking calculations

### Integration Tests
- End-to-end SFTP workflows
- TUI state transitions
- Error handling scenarios
- Performance with large files

### User Acceptance Testing
- Usability testing with target users
- Performance testing with real media servers
- Security testing for authentication flows
- Compatibility testing with various SFTP servers

## Success Metrics

### Performance
- Connection establishment time < 5 seconds
- Directory listing time < 2 seconds for 1000 files
- File transfer rate matching network capacity
- Memory usage < 500MB for 100MB files

### User Experience
- Task completion time reduction of 40% compared to manual process
- User satisfaction rating > 4.0/5.0
- Error rate < 1% for typical operations
- Learning curve < 15 minutes for basic tasks

## Dependencies

### External Libraries
- `github.com/pkg/sftp` for SFTP operations
- `golang.org/x/crypto/ssh` for SSH authentication
- `github.com/charmbracelet/bubbletea` for TUI framework
- `github.com/charmbracelet/bubbles` for UI components

### System Requirements
- Go 1.19 or higher
- SSH key pair in standard location
- Network access to remote SFTP server
- Sufficient disk space for temporary files

## Future Enhancements

### Planned Features
- Support for additional authentication methods
- Integration with cloud storage services
- Advanced image filtering (by date, camera model, etc.)
- Scheduling for automated optimization tasks
- Web-based dashboard for monitoring operations

### Potential Improvements
- Machine learning-based optimization suggestions
- Parallel processing for multiple files
- Integration with image recognition for content-aware optimization
- Mobile app companion for remote management

This PRD provides a comprehensive foundation for implementing the SFTP extension to Photoptim using Bubble Tea for the TUI. The design focuses on creating an intuitive, efficient, and secure user experience while leveraging the power of Bubble Tea's component system for a responsive terminal interface.

---

# Refined Technical Design (Aug 2025 Alignment)

The following sections refine and lock the implementation details agreed during design review discussions. They supersede or extend earlier high‑level notes.

## High-Level Architecture

New / refactored internal packages:

| Package | Purpose |
|---------|---------|
| `internal/remotefs` | Protocol-agnostic filesystem abstraction (`RemoteFS`, path sanitation, symlink cycle guard) |
| `internal/sftp` | Concrete SFTP implementation using `golang.org/x/crypto/ssh` + `github.com/pkg/sftp` |
| `internal/cache` | bbolt-backed directory metadata cache (TTL-based) |
| `internal/optimizer` | Refactored optimizer behind `Optimizer` interface; adds PNG (lossless via `bimg`), tiny/unsupported skip logic |
| `internal/pipeline` | Orchestrates per-file end‑to‑end workflow (download → optimize → upload) with progress events |
| `internal/progress` | Byte-level progress instrumentation (readers/writers + aggregate) |
| `internal/audit` | Optional JSON audit log (file size delta + outcome) |
| `internal/metrics` | In-memory counters & durations (no persistence) |
| `internal/config` | Flags/env/config file merging; known_hosts & settings persistence |

TUI additions extend existing `internal/tui` with new state structs (Connection, Browser, Selection, Config, Progress, Results) and components (file list w/ filtering, progress dashboard, retry/actions pane).

## Core Interfaces

```go
// remote filesystem abstraction
type RemoteFS interface {
   Connect(ctx context.Context, cfg ConnectionConfig) error
   Close() error
   List(ctx context.Context, path string) ([]RemoteEntry, error)
   Stat(ctx context.Context, path string) (RemoteEntry, error)
   Open(ctx context.Context, path string) (io.ReadCloser, RemoteEntry, error)
   Create(ctx context.Context, path string, overwrite bool) (io.WriteCloser, error)
   Join(elem ...string) string
   Root() string
}

type RemoteEntry struct {
   Path       string
   Name       string
   Size       int64
   Mode       fs.FileMode
   ModTime    time.Time
   IsDir      bool
   Symlink    bool
   TargetPath string // resolved target if symlink (within chroot)
}

// image optimization abstraction
type Optimizer interface {
   Optimize(ctx context.Context, in []byte, format string, params OptimizeParams) (out []byte, res OptimizeResult, err error)
}

type OptimizeParams struct {
   JPEGQuality int
}

type OptimizeResult struct {
   OriginalSize  int64
   OptimizedSize int64
   Skipped       bool
   Reason        string // e.g. "tiny-image", "unsupported-format"
   Duration      time.Duration
}

// streaming progress events
type Phase string
const (
   PhaseDownload Phase = "download"
   PhaseOptimize Phase = "optimize"
   PhaseUpload   Phase = "upload"
)

type ProgressEvent struct {
   FileID    int
   Name      string
   Phase     Phase
   Bytes     int64
   Total     int64
   Done      bool
   Err       error
   Timestamp time.Time
}

// audit logging
type AuditRecord struct {
   Path           string
   OriginalSize   int64
   OptimizedSize  int64
   SavingsBytes   int64
   SavingsPercent float64
   DurationMs     int64
   Status         string // success|skipped|failed
   Reason         string
}
```

## Invocation Modes

New CLI subcommand: `photoptim sftp`.

Modes:
1. Interactive (default) – launches Bubble Tea SFTP workflow.
2. Batch (`--batch`) – non-TUI execution with structured summary and exit codes.

## CLI Flags & Environment

| Flag | Description | Notes |
|------|-------------|-------|
| `--host` | SFTP host | required (batch) |
| `--port` | Port (default 22) | optional |
| `--user` | Username | required (batch) |
| `--remote-path` | Initial (chroot) path | required (batch) |
| `--key` | Path to private key | optional (agent/key/password fallback) |
| `--password` | Password auth fallback | optional |
| `--quality` | JPEG quality (1–100) | affects JPEG only |
| `--size-threshold` | Inclusive size filter (e.g. 500KB, 2MB) | live-refilter TUI (400ms debounce) |
| `--concurrency` | Worker concurrency | default 4 (overrides env) |
| `--ttl` | Directory cache TTL override | default 2m |
| `--save-config` | Persist connection & quality settings | optional YAML |
| `--verbose` | Enable verbose logs | toggle inside TUI (key V) |
| `--audit` | Enable audit JSON emission | file path printed |
| `--skip-cache` | Force fresh listing | bypass cache |
| `--keep-temp` | Retain temp dirs post run | default remove on success |
| `--batch` | Run non-interactive | implies required flags |

Environment:
* `PHOTOPTIM_CONCURRENCY` – fallback for concurrency when no flag provided.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 3 | Zero optimizable files found |
| 4 | Partial failures after retries |
| 5 | Connection / authentication failure |
| 6 | Internal pipeline error |

## Authentication & Security

Order of attempts: SSH agent → specified key (supports RSA, ED25519, ECDSA; passphrase prompt if needed) → password (if provided). Host key verification prompts user on first connect; fingerprint stored at `$XDG_CONFIG_HOME/photoptim/known_hosts` (fallback `~/.config/photoptim/known_hosts`). Four auto-reconnect attempts with exponential backoff (e.g. 250ms, 500ms, 1s, 2s).

No encryption for temp files (explicitly deferred). Audit log limited to size deltas & status. No external telemetry; only in-memory metrics.

## Browsing & Selection Behavior

* Chroot enforced to initial `--remote-path` (cannot navigate above).
* Symlinks followed if target remains within chroot; cycle detection via visited inode/path set.
* Hidden (dot) files excluded.
* Directory listing loads all entries (capped 1000). If >1000 entries, user must refine size filter before selection is enabled.
* Supported extensions: `.jpg`, `.jpeg`, `.png` (case-insensitive). Unsupported are displayed (optional) but automatically skipped at optimization stage with clear reason.
* No recursive multi-directory selection.

## Filtering Logic

* Size threshold inclusive `>=`.
* Human-friendly units accepted: KB / MB (binary 1024 scale).
* Live re-filter in TUI using 400ms debounce after last change.
* Tiny image skip threshold: <15 KB (prevents negligible savings).

## Temporary Storage Layout

Session temp root: e.g. `$TMPDIR/photoptim_sftp_<timestamp>/` containing:
```
orig/        # downloaded originals (mirrors relative structure)
optimized/   # optimized results
audit.json   # (optional) audit log when enabled
```
On hard cancel: partial files removed; completed originals/optimized retained unless not yet uploaded. `--keep-temp` preserves entire tree.

## Optimization Strategy

JPEG: Quality slider (default from existing tool or 80 if unset). PNG: lossless recompression via `bimg` (libvips) – if libvips unavailable, PNG optimization gracefully skips with explanatory reason. Unsupported formats & tiny images marked skipped.

Concurrency: Each worker performs full per-file pipeline sequentially (download → optimize → upload). Default 4; override via flag/env. Memory buffering acceptable for target 0.5–5 MB typical file size range.

## Pipeline & Progress

Per-file pipeline emits `ProgressEvent` for each phase with byte-level updates (download/upload). Optimization phase provides synthetic total = original size, updating bytes once at completion (or progressive if future streaming optimization added). Overall progress aggregates per-file totals (original size sum) to compute cumulative percentage.

Cancellation (Ctrl+X): Immediate context cancellation; in-flight transfers aborted; partial temp artifacts cleaned. Log summarises processed / failed / pending counts.

## Caching Strategy

* Backend: bbolt single DB file (e.g. `$XDG_CACHE_HOME/photoptim/cache.db`).
* Bucket per directory: key = normalized path; value = JSON (entries + `storedAt` timestamp).
* TTL default 2 minutes; stale entries re-fetched automatically; user can force refresh (Ctrl+L / `--skip-cache`).
* Cached data persisted across runs to speed repeated navigation; no file content caching (metadata only).

## Audit & Metrics

Audit (when `--audit`): JSON array (streamed append-safe) of `AuditRecord`. Fields: original size, optimized size, savings bytes, percent, duration, status, reason. Metrics: in-memory counters (files processed, skipped, failed; cumulative bytes saved; durations by phase) displayed in Results screen.

## Error Handling & Messaging

Errors shown in dual format: user-friendly summary + technical detail (expandable when verbose). Retry UI allows per-file retry or retry-all. Network interruptions trigger up to 4 reconnect attempts automatically; if exhausted, affected files marked failed with retry option.

## Logging & Verbosity

`--verbose` enables technical logs (SFTP operations, cache hits/misses, retry attempts). In TUI, key `V` toggles verbose pane. Non-verbose mode shows concise status lines only.

## Keybindings (Additions)

| Key | Context | Action |
|-----|---------|--------|
| Ctrl+L | Global | Refresh / invalidate cache |
| Ctrl+X | Progress | Hard cancel pipeline |
| V | Any (TUI) | Toggle verbose log pane |
| R | Browser (selected file) | Rename (handle unusual chars) |
| Space | Browser/Selection | Toggle selection |
| A | Selection | Select all (within filtered set, <=1000) |
| N | Selection | Deselect all |

Existing global keys (quit etc.) remain unchanged.

## Results & Retry Screen

Displays table: Name | Original | Optimized | Savings % | Status. Actions: (Retry Failed, Retry Selected, Export Audit Path, Return to Browser). Failed entries highlight reason (hover/expand for technical detail).

## Unicode & Unusual Filenames

Names containing spaces or non-printable ASCII (outside 32–126) flagged with ⚠. Tooltip offers rename (`R`). Rename only affects local temp & upload target (server file overwritten under new chosen name if different; original remote name preserved only if unchanged).

## Disk Space Pre-Check

Phase 4 enhancement (optional): compute required space = sum(original sizes (filtered) * 2) for worst-case plus overhead; warn if insufficient. Not critical path for initial release.

## Dependencies (Additions)

| Dependency | Purpose |
|------------|---------|
| `github.com/pkg/sftp` | SFTP operations |
| `golang.org/x/crypto/ssh` | SSH auth + host key handling |
| `go.etcd.io/bbolt` | Persistent metadata cache |
| `github.com/h2non/bimg` | PNG lossless recompression (libvips) |

## Performance Targets (Confirmed)

* Typical file size: 0.5–5 MB
* Median directory: ~200 files (still efficient well below 1000 cap)
* Directory listing (<=1000 entries) < 2s
* Connection setup < 5s
* Memory usage < 500MB (headroom ample under target sizes)

## Testing Strategy

Unit Tests:
* Optimizer (JPEG quality application, PNG skip when libvips missing, tiny-image skip, unsupported format skip)
* Cache TTL logic and refresh path
* Host key acceptance & persistence
* Progress event aggregation (percent correctness)
* Path chroot & symlink cycle prevention

Integration Tests (using mock `RemoteFS`):
* End-to-end batch mode (success, zero-files, partial failures)
* Retry logic (inject transient errors)
* Cancellation mid-transfer

Mocking:
* `RemoteFS` in-memory implementation simulating network latency + failures
* Fixture generator for directory entries & test image blobs

Exit Code Assertions:
* Validate each documented exit condition via controlled scenarios.

## Phased Implementation (Revised)

1. Interfaces + optimizer refactor
2. SFTP client (auth, host key prompt, basic ops)
3. Cache layer (bbolt) + TTL logic
4. Pipeline + progress events + audit & metrics scaffolding
5. CLI subcommand & batch mode (exit codes, summary)
6. TUI states (Connection → Results) baseline navigation
7. Advanced UX (debounced filter, >1000 gating, rename, verbose pane, retry UI)
8. PNG optimization integration (`bimg`) + tiny skip logic
9. Documentation (`SFTP_USAGE.md`, update `TUI_USAGE.md`, architecture diagram optional)
10. Extended tests, polish, performance tuning, optional disk space pre-check

## Out of Scope (Initial Release)

* Encryption of temp files
* Recursive directory selection
* Streaming optimization for very large files
* Protocols beyond SFTP
* External telemetry / analytics persistence

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| libvips absent for PNG | PNG optimization skipped | Detect & warn once; mark skipped with reason |
| Large directories >1000 entries | Performance/UI overwhelm | Enforced gating & user filter refinement |
| Network flakiness | Partial failures | 4 reconnect attempts + retry UI |
| Host key spoofing risk | Security compromise | Explicit fingerprint display & user confirmation |
| Memory spikes (many concurrent files) | Resource pressure | Concurrency capped; per-file sequential pipeline |

## Audit & Reporting Examples

Example audit record (JSON):
```json
{
  "Path": "images/sample.jpg",
  "OriginalSize": 2457600,
  "OptimizedSize": 1536000,
  "SavingsBytes": 921600,
  "SavingsPercent": 37.5,
  "DurationMs": 842,
  "Status": "success",
  "Reason": ""
}
```

## Open Extension Points

* `RemoteFS` ready for additional protocols (S3, WebDAV) via new implementations.
* `Optimizer` can add WebP/AVIF in future (extend params with per-format settings).
* Metrics struct can later emit to external sinks (feature-flagged).

---

This refined section finalizes the SFTP extension blueprint and should guide implementation, code review, and testing criteria.