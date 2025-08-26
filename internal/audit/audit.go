package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// Record holds audit information for a processed file.
type Record struct {
	Path           string  `json:"path"`
	OriginalSize   int64   `json:"originalSize"`
	OptimizedSize  int64   `json:"optimizedSize"`
	SavingsBytes   int64   `json:"savingsBytes"`
	SavingsPercent float64 `json:"savingsPercent"`
	Status         string  `json:"status"`
	Reason         string  `json:"reason,omitempty"`
}

// Logger appends records to a JSON file (array) in a simplistic manner.
type Logger struct {
	mu    sync.Mutex
	path  string
	first bool
	file  *os.File
}

func NewLogger(path string) (*Logger, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if _, err := f.WriteString("["); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &Logger{path: path, file: f, first: true}, nil
}

func (l *Logger) Append(r Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	if !l.first {
		if _, err := l.file.WriteString(","); err != nil {
			return err
		}
	}
	b, _ := json.Marshal(r)
	if _, err := l.file.Write(b); err != nil {
		return err
	}
	l.first = false
	return nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	_, _ = l.file.WriteString("]\n")
	return l.file.Close()
}
