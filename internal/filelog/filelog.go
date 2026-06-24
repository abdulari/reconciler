package filelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RotatingWriter struct {
	dir      string
	prefix   string
	mu       sync.Mutex
	file     *os.File
	curDate  string
}

func New(dir, prefix string) *RotatingWriter {
	return &RotatingWriter{dir: dir, prefix: prefix}
}

func (w *RotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("20060102")
	if today != w.curDate {
		if w.file != nil {
			w.file.Close()
		}
		name := filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.prefix, today))
		f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		w.file = f
		w.curDate = today
	}

	return w.file.Write(p)
}

func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}