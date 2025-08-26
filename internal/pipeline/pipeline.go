package pipeline

import (
	"context"
	"errors"
	"sync"
	"time"

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
	Concurrency int
}

func (o *Orchestrator) Run(ctx context.Context, tasks []FileTask) (<-chan ProgressEvent, <-chan error) {
	prog := make(chan ProgressEvent)
	errs := make(chan error, 1)
	go func() {
		defer close(prog)
		defer close(errs)
		if o.Concurrency <= 0 {
			errs <- errors.New("invalid concurrency")
			return
		}
		var wg sync.WaitGroup
		sem := make(chan struct{}, o.Concurrency)
		for i, t := range tasks {
			i, t := i, t
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				// Placeholder: emit a synthetic completed progress event
				prog <- ProgressEvent{FileID: i, Name: t.Entry.Name, Phase: PhaseDownload, Bytes: t.Entry.Size, Total: t.Entry.Size, Done: true, Timestamp: time.Now()}
			}()
		}
		wg.Wait()
	}()
	return prog, errs
}
