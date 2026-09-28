package dataset

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/preprocessing"
)

// Rutas por defecto: las mismas que usan main.go, cmd/train y cmd/benchmark.
const (
	DefaultRaw     = "data/raw/" + FileName
	DefaultClean   = "data/processed/paysim_clean.csv"
	DefaultSummary = "data/processed/resumen_limpieza.json"
)

// EnsureRaw deja listo el CSV crudo: si no existe y es el archivo de PaySim,
// lo descarga desde Kaggle en su carpeta (como una celda de Colab que baja el
// dataset al ejecutarse). Un archivo con otro nombre que no existe es error.
func EnsureRaw(path string, progress io.Writer) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if filepath.Base(path) != FileName {
		return fmt.Errorf("no existe el archivo %s", path)
	}
	fmt.Fprintf(progress, "No se encontro %s: descargando PaySim desde Kaggle (una sola vez)...\n", path)
	_, err := Download(KaggleURL, filepath.Dir(path), progress)
	return err
}

// EnsureClean deja listo el CSV limpio: si no existe, asegura el crudo y
// corre la limpieza concurrente con numWorkers. Si ya existe no hace nada.
func EnsureClean(rawPath, cleanPath, summaryPath string, numWorkers int, progress io.Writer) error {
	if _, err := os.Stat(cleanPath); err == nil {
		return nil
	}
	if err := EnsureRaw(rawPath, progress); err != nil {
		return err
	}
	for _, path := range []string{cleanPath, summaryPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	fmt.Fprintf(progress, "No se encontro %s: limpiando el dataset con %d workers...\n", cleanPath, numWorkers)
	summary, err := preprocessing.Run(rawPath, cleanPath, summaryPath, numWorkers)
	if err != nil {
		os.Remove(cleanPath) // no dejar un CSV limpio a medias
		return fmt.Errorf("limpiar dataset: %w", err)
	}
	fmt.Fprintf(progress, "Limpieza lista: %d filas validas de %d en %.1fs\n",
		summary.Valid, summary.TotalRead, summary.ElapsedSeconds)
	return nil
}
