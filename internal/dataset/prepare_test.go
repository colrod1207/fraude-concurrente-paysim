package dataset

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/preprocessing"
)

func TestEnsureRaw_ExistingFileIsNoOp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mi_dataset.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureRaw(path, io.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureRaw_MissingCustomFileIsError(t *testing.T) {
	// Solo se descarga si se pide el archivo de PaySim: un nombre distinto
	// que no existe es un error de ruta del usuario, no algo que bajar.
	err := EnsureRaw(filepath.Join(t.TempDir(), "otro.csv"), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "otro.csv") {
		t.Fatalf("se esperaba un error que nombre el archivo, got %v", err)
	}
}

func TestEnsureClean_RunsCleaningOnceAndCreatesFolders(t *testing.T) {
	dir := t.TempDir()
	raw := filepath.Join(dir, "raw.csv")
	if err := preprocessing.WriteSyntheticCSV(raw); err != nil {
		t.Fatal(err)
	}
	clean := filepath.Join(dir, "processed", "clean.csv") // carpeta inexistente
	summary := filepath.Join(dir, "processed", "resumen.json")

	if err := EnsureClean(raw, clean, summary, 2, io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(clean)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(string(data), "\n"); lines != 8 { // header + 7 filas validas
		t.Fatalf("el CSV limpio tiene %d lineas, want 8", lines)
	}
	if _, err := os.Stat(summary); err != nil {
		t.Fatalf("falta el resumen: %v", err)
	}

	// Si el CSV limpio ya existe no se vuelve a limpiar (ni se necesita el crudo).
	if err := os.Remove(raw); err != nil {
		t.Fatal(err)
	}
	if err := EnsureClean(raw, clean, summary, 2, io.Discard); err != nil {
		t.Fatalf("no debio volver a limpiar: %v", err)
	}
}
