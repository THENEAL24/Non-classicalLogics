package fuzzy

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Journal — потокобезопасный журнал операций в файл и для GUI.
type Journal struct {
	mu   sync.Mutex
	path string
	lines []string
}

func NewJournal(path string) *Journal {
	return &Journal{path: path, lines: make([]string, 0, 64)}
}

func (j *Journal) Append(format string, args ...any) {
	line := fmt.Sprintf("[%s] %s",
		time.Now().Format("2006-01-02 15:04:05"),
		fmt.Sprintf(format, args...))
	j.mu.Lock()
	defer j.mu.Unlock()
	j.lines = append(j.lines, line)
	if j.path != "" {
		_ = os.MkdirAll(filepath.Dir(j.path), 0755)
		f, err := os.OpenFile(j.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			_, _ = fmt.Fprintln(f, line)
			_ = f.Close()
		}
	}
}

func (j *Journal) Text() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return stringsJoin(j.lines)
}

func stringsJoin(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	out := lines[0]
	for i := 1; i < len(lines); i++ {
		out += "\n" + lines[i]
	}
	return out
}
