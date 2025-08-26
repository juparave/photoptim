package pipeline

import (
	"context"
	"io"
	"sync"
	"time"

	"photoptim/internal/optimizer"
	"photoptim/internal/remotefs"
)

type Phase string

const (
	PhaseDownload Phase = "download"
	PhaseOptimize Phase = "optimize"
	PhaseUpload   Phase = "upload"
)

// ProgressEvent describes streaming progress.
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

// FileTask describes an optimization/upload task.
type FileTask struct {
	Entry remotefs.RemoteEntry
}

// Orchestrator manages pipeline execution (scaffold only).
type Orchestrator struct {
	FS            remotefs.RemoteFS
	Opt           optimizer.Optimizer
	Concurrency   int
	JPEGQuality   int
	TinyThreshold int64
}

func (o *Orchestrator) Run(ctx context.Context, tasks []FileTask) (<-chan ProgressEvent, <-chan error) {
	prog := make(chan ProgressEvent)
	errs := make(chan error, 1)
	if o.Concurrency <= 0 {
		o.Concurrency = 1
	}
	if o.TinyThreshold == 0 {
		o.TinyThreshold = 15 * 1024
	}
	go func() {
		defer close(prog)
		defer close(errs)
		sem := make(chan struct{}, o.Concurrency)
		var wg sync.WaitGroup
		for i, task := range tasks {
			i, task := i, task
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				select {
				case <-ctx.Done():
					return
				default:
				}
				rc, entry, err := o.FS.Open(ctx, task.Entry.Path)
				if err != nil {
					prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseDownload, Err: err, Done: true, Timestamp: time.Now()}
					return
				}
				data, err := io.ReadAll(rc)
				_ = rc.Close()
				if err != nil {
					prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseDownload, Err: err, Done: true, Timestamp: time.Now()}
					return
				}
				prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseDownload, Bytes: int64(len(data)), Total: entry.Size, Done: true, Timestamp: time.Now()}
				// optimize
				out, res, optErr := o.Opt.OptimizeBytes(data, detectFormat(task.Entry.Name), optimizer.Params{JPEGQuality: o.JPEGQuality})
				if res.Skipped || optErr != nil {
					prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseOptimize, Bytes: res.OriginalSize, Total: res.OriginalSize, Done: true, Err: optErr, Timestamp: time.Now()}
					return
				}
				prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseOptimize, Bytes: res.OriginalSize, Total: res.OriginalSize, Done: true, Timestamp: time.Now()}
				wc, err := o.FS.Create(ctx, task.Entry.Path, true)
				if err != nil {
					prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseUpload, Done: true, Err: err, Timestamp: time.Now()}
					return
				}
				if _, err := wc.Write(out); err != nil {
					_ = wc.Close()
					prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseUpload, Done: true, Err: err, Timestamp: time.Now()}
					return
				}
				_ = wc.Close()
				prog <- ProgressEvent{FileID: i, Name: task.Entry.Name, Phase: PhaseUpload, Bytes: int64(len(out)), Total: int64(len(out)), Done: true, Timestamp: time.Now()}
			}()
		}
		wg.Wait()
	}()
	return prog, errs
}

func detectFormat(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i+1:]
		}
	}
	return ""
}
