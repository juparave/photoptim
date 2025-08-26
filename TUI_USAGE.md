## Photoptim TUI Usage Guide

This guide describes how to use the local file optimization TUI and the new SFTP workflow.

---

## Local File Optimization TUI

### Launch

```
./photoptim-tui
```

### Workflow
1. File Picker Screen
	- Arrow keys navigate directories / files.
	- Enter selects an image (e.g. `testdata/sample.jpg`).
	- Allowed extensions: `.jpg`, `.jpeg`, `.png`.
2. Optimization Mode Selection
	- Choose: Single Image Optimization or Batch Optimization.
	- Enter to confirm.
3. Quality Input
	- Enter JPEG quality (1–100). Default 80.
4. Output Directory
	- Provide output directory name (e.g. `optimized`). Created if missing.
5. Optimization Progress
	- Progress bar updates.
	- Status messages show current step.
	- Press `q` (or `ctrl+c`) to quit after completion.

### Keybindings (Local Mode)
| Key | Action |
|-----|--------|
| ↑/↓ / ←/→ | Navigate / move focus |
| Enter | Confirm selection / next step |
| Esc | Go back one step |
| q / Ctrl+C | Quit |

---

## SFTP Optimization TUI (New)

The SFTP workflow lets you connect to a remote server, browse images, optimize them locally, and upload optimized versions back (overwriting originals unless skipped/failed). Requires an SSH-accessible host and key/password credentials.

### Launch

```
photoptim sftp
```

### Connection Screen
Fields:
| Field | Description |
|-------|-------------|
| Host | Remote hostname or IP |
| Port | Defaults to 22 |
| User | SSH username |
| Remote Path | Initial working directory (chroot root) |
| Key (optional) | Path to private key; agent used automatically if available |
| Password (fallback) | Only used if agent / key auth fails |

Actions:
* Enter: attempt connection.
* On first connect to unknown host, fingerprint prompt is shown (accept / reject).

### Browser Screen
Shows remote directory (chrooted to initial path):
* Hidden (dot) files excluded.
* Symlinks followed only if target remains within chroot.
* Hard cap: if >1000 entries, a gating message appears—apply a size filter to reduce set.

Columns:
| Name | Size | Type | Modified |

Filtering:
* Size threshold input (supports values like `500KB`, `2MB`). Inclusive (>=).
* 400ms debounce before list refilters.

Selection:
* Space: toggle file selection.
* A: select all (current filtered set <=1000).
* N: deselect all.
* ⚠ indicates unusual filename (spaces or non-ASCII); press `R` to rename before optimization/upload.

### Selection Screen
* Displays chosen files count.
* Proceed to configuration (quality / concurrency) or adjust selection.

### Optimization Configuration Screen
| Setting | Description |
|---------|-------------|
| JPEG Quality | Slider/input 1–100 (PNG uses lossless recompress) |
| Concurrency | Worker count (default 4, env `PHOTOPTIM_CONCURRENCY`) |
| Size Threshold | Optional override applied retroactively |
| Keep Temp | Retain temp dirs after finishing |
| Audit Log | Enable JSON audit (savings per file) |

### Progress Screen
Displays:
* Overall progress bar.
* Per-file stacked bars (Download / Optimize / Upload) with byte-level progress for download/upload.
* Counters: Completed | Skipped | Failed | Savings.
* Verbose log pane (toggle `V`).

Cancellation:
* Ctrl+X: hard cancel (in-flight file aborted; partial temp deleted). Summary shown.

### Results Screen
* Table: Name | Original | Optimized | Savings % | Status.
* Actions: Retry Failed, Retry Selected, Export Audit Path, Return to Browser.
* Failed entries show reason (hover/expand in verbose mode).

### Batch Mode (Non-TUI)
Run directly via flags for automation:
```
photoptim sftp --batch \
  --host example.com --user alice --remote-path /photos \
  --quality 80 --size-threshold 500KB --concurrency 6 --audit
```
Exit codes:
| Code | Meaning |
|------|---------|
| 0 | Success |
| 3 | Zero optimizable files |
| 4 | Partial failures |
| 5 | Connection/auth failure |
| 6 | Internal pipeline error |

### Keybindings (SFTP Mode Summary)
| Key | Action |
|-----|--------|
| Arrow Keys | Navigate lists / fields |
| Enter | Confirm / advance |
| Esc | Back |
| Space | Toggle selection |
| A | Select all (<=1000) |
| N | Deselect all |
| R | Rename highlighted file |
| V | Toggle verbose log pane |
| Ctrl+L | Refresh / invalidate cache |
| Ctrl+X | Cancel processing (Progress screen) |
| Ctrl+C / q | Quit |

### Notes
* Tiny images (<15KB) skipped automatically (reported in results).
* Unsupported formats skipped with reason.
* PNG optimization requires libvips (`bimg`); absence results in skip with warning.
* Host fingerprints stored under `$XDG_CONFIG_HOME/photoptim/known_hosts`.
* Cache TTL default 2 minutes; refresh with Ctrl+L.

---

## Troubleshooting
| Symptom | Resolution |
|---------|------------|
| PNGs all skipped | Install libvips; reinstall or rebuild binary with CGO enabled if needed |
| Large directory gated | Apply a size threshold or navigate deeper to reduce entries ≤1000 |
| Connection fails | Verify SSH key permissions and host/port; check known_hosts fingerprint changes |
| No files optimized (exit code 3) | Adjust size threshold or ensure supported extensions present |

---

The TUI provides an interactive way to optimize images locally and remotely without memorizing detailed CLI options.