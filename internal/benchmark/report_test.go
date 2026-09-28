package benchmark

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"
)

var sampleResults = []Result{
	{Task: "limpieza", Mode: "secuencial", Workers: 1,
		Runs: []time.Duration{2 * time.Second, 2 * time.Second}, Mean: 2 * time.Second,
		Speedup: 1, Efficiency: 1, Profile: Profile{CPUPercent: 12.5, CoresUsed: 1, PeakHeapMB: 10, AllocMB: 300}},
	{Task: "limpieza", Mode: "concurrente", Workers: 4,
		Runs: []time.Duration{time.Second, 500 * time.Millisecond}, Mean: 750 * time.Millisecond,
		Speedup: 2.6667, Efficiency: 0.6667, Profile: Profile{CPUPercent: 40, CoresUsed: 3.2, PeakHeapMB: 12, AllocMB: 310}},
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, sampleResults); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || len(rows[0]) != len(csvHeader) {
		t.Fatalf("se esperaban header + 2 filas de %d columnas, got %v", len(csvHeader), rows)
	}
	want := []string{"limpieza", "concurrente", "4", "2", "0.7500", "2.667", "0.667",
		"40.0", "3.20", "12.0", "310.0", "1.0000;0.5000"}
	for i, value := range want {
		if rows[2][i] != value {
			t.Errorf("columna %s = %q, want %q", csvHeader[i], rows[2][i], value)
		}
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, sampleResults); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("se esperaban 4 lineas (header, separador, 2 filas), got %d:\n%s", len(lines), buf.String())
	}
	if want := "| limpieza | concurrente | 4 | 0.7500 | 2.67 | 0.67 | 40.0 | 3.20 | 12.0 | 310.0 |"; lines[3] != want {
		t.Fatalf("fila = %q\nwant  %q", lines[3], want)
	}
}
