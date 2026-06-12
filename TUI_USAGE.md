# Photoptim TUI Usage Guide

This guide describes the two interactive terminal interfaces: the **local** file
optimization TUI and the **SFTP** remote optimization TUI. It reflects the
behavior actually implemented in the code.

---

## Local File Optimization TUI

### Launch

```
photoptim-tui
```

### Workflow
1. **File Picker** — browse the current directory tree and select images.
2. **Mode Selection** — choose the optimization mode (Batch Optimization).
3. **Quality Input** — enter JPEG quality (1–100, default 80).
4. **Output Directory** — name of the output directory (created if missing).
5. **Progress** — a progress bar advances per file, and the last few results are
   listed. A summary (`N optimized, M failed`) is shown on completion.

Supported input formats: `.jpg`, `.jpeg`, `.png`. Files with any other
extension are skipped and reported as `⊘ unsupported format`.

### Keybindings (Local Mode)
| Key | Action |
|-----|--------|
| ↑/k, ↓/j | Move up / down |
| Enter | Open directory / toggle file selection |
| Space | Toggle file selection |
| a | Select all files in the current directory |
| c | Clear selection |
| s | Proceed to optimization |
| Esc / Backspace | Go back one step / up a directory |
| q / Ctrl+C | Quit |

> Note: while typing in the quality or output-directory fields, `q` is treated
> as text; use Ctrl+C to quit from those screens.

---

## SFTP Optimization TUI

The SFTP workflow connects to a remote server, browses images, optimizes them,
and writes the optimized version back **in place** (atomically — see below).

### Launch

```
photoptim sftp
```

### Connection Screen
Fields (navigate with Tab / Shift+Tab / ↑ / ↓):

| Field | Description |
|-------|-------------|
| Host | Remote hostname or IP |
| Port | Defaults to 22 |
| User | SSH username |
| Pass | Password, or passphrase for an encrypted key (fallback) |
| Key  | Path to a private key (e.g. `~/.ssh/id_rsa`); optional |
| Path | Initial remote directory; blank uses the login home directory |

Authentication is attempted in this order:
1. **ssh-agent** (if `SSH_AUTH_SOCK` is set)
2. **Identity files** — the `Key` field if provided, otherwise the common
   defaults `~/.ssh/id_ed25519`, `~/.ssh/id_rsa`, `~/.ssh/id_ecdsa`
3. **Password** (fallback)

Move focus to `[ Connect ]` and press Enter to connect.

#### Host key verification
Server host keys are verified against
`$XDG_CONFIG_HOME/photoptim/known_hosts` (default `~/.config/photoptim/known_hosts`):

- **First connection to a host** — trust on first use: the key is recorded and
  accepted automatically.
- **Subsequent connections** — the key must match the recorded one.
- **Changed key** — the connection is rejected (possible man-in-the-middle). If
  the change is expected, remove the stale line for that host from the
  `known_hosts` file and reconnect.

### Browser Screen
Lists the current remote directory (dot files are hidden). Directories are shown
first; files ≥ 10 MB are highlighted. The footer shows the current Quality and
Resize settings.

| Key | Action |
|-----|--------|
| ↑ / ↓ | Move selection |
| Enter | Open directory, or (if files are selected) start optimization |
| Space | Toggle file selection |
| a | Select all files in the current directory |
| c | Clear selection |
| +/- | Increase / decrease JPEG quality (steps of 5, 1–100) |
| r | Cycle resize preset (Disabled → device presets → …) |
| s | Toggle sort (name ⇄ size) |
| Backspace / ← | Go up one directory |
| q / Ctrl+C | Quit |

Resize presets include Disabled, iPhone / Samsung / Pixel / iPad sizes, and
Full HD / 2K / 4K. When a preset other than Disabled is active, images are
resized (preserving aspect ratio) before re-encoding.

### Optimization
On Enter with one or more files selected, each selected file is processed
sequentially:

1. Downloaded into memory.
2. Re-encoded (JPEG at the chosen quality; PNG re-encoded losslessly) and
   optionally resized.
3. Written back to the **same remote path**.

Remote writes are **atomic**: the optimized data is written to a temporary file
alongside the original and then renamed over it on success. If the write or
rename fails, the temporary file is removed and the **original is left
unchanged**.

A file is skipped (and reported) when:
- its extension is not `.jpg`, `.jpeg`, or `.png`, or
- the optimized result would not be smaller than the original
  (`original is already optimal`).

Progress is shown with an overall bar and the most recent results. On
completion a summary (`N optimized, M failed`) is displayed and the listing is
refreshed.

### Batch Mode (Non-Interactive)
> **Status:** not yet implemented. `photoptim sftp --batch --host … --user …`
> currently only verifies connectivity and exits; it does not run the
> optimization pipeline. Use the interactive TUI for remote optimization.

---

## Troubleshooting
| Symptom | Resolution |
|---------|------------|
| `host key mismatch` on connect | If the change is expected, remove the host's line from `~/.config/photoptim/known_hosts` and reconnect. |
| `no auth methods available` | Start ssh-agent, ensure a default key exists in `~/.ssh`, or supply a key path / password. |
| Connection fails | Verify host, port, username, and key permissions. |
| File reported `already optimal` | The current encoding is already at or below the target size; nothing is written. |
| Format skipped | Only `.jpg`, `.jpeg`, and `.png` are supported. |

---

The TUI provides an interactive way to optimize images locally and remotely
without memorizing CLI options.
