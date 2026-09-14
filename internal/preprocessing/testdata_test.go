package preprocessing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSyntheticCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clean_input.csv")
	if err := WriteSyntheticCSV(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 14 { // 1 header + 13 filas de datos
		t.Fatalf("expected 14 lines (header + 13 rows), got %d", len(lines))
	}
}
