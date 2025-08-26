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