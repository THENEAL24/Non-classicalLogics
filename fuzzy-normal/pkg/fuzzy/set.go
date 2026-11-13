package fuzzy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Set — множество нечётких величин.
type Set struct {
	Name   string        `json:"name"`
	Values []NormalFuzzy `json:"values"`
}

func NewSet(name string, values ...NormalFuzzy) Set {
	return Set{Name: name, Values: append([]NormalFuzzy(nil), values...)}
}

func (s Set) Len() int { return len(s.Values) }

func (s Set) Clone() Set {
	return Set{Name: s.Name, Values: append([]NormalFuzzy(nil), s.Values...)}
}

// Append добавляет величины в множество.
func (s *Set) Append(v ...NormalFuzzy) {
	s.Values = append(s.Values, v...)
}

// SaveJSON сохраняет множество в файл.
func (s Set) SaveJSON(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadJSON загружает множество из файла.
func LoadJSON(path string) (Set, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Set{}, err
	}
	var s Set
	if err := json.Unmarshal(data, &s); err != nil {
		return Set{}, err
	}
	return s, nil
}

// FormatTable — текстовое представление для экрана и лога.
func (s Set) FormatTable(maxRows int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Множество «%s», элементов: %d\n", s.Name, len(s.Values))
	if maxRows <= 0 {
		maxRows = len(s.Values)
	}
	n := maxRows
	if n > len(s.Values) {
		n = len(s.Values)
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "  [%d] %s\n", i+1, s.Values[i])
	}
	if n < len(s.Values) {
		fmt.Fprintf(&b, "  ... ещё %d\n", len(s.Values)-n)
	}
	return b.String()
}
