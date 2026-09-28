// Ejecuta la limpieza concurrente del dataset PaySim (PC1, CC65).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/dataset"
	"github.com/colrod1207/fraude-concurrente-paysim/internal/preprocessing"
)

func main() {
	input := flag.String("input", dataset.DefaultRaw, "ruta del CSV crudo de PaySim (se descarga si falta)")
	output := flag.String("output", dataset.DefaultClean, "ruta del CSV limpio de salida")
	summaryPath := flag.String("summary", dataset.DefaultSummary, "ruta del resumen JSON")
	workers := flag.Int("workers", 4, "numero de goroutines worker")
	flag.Parse()

	if err := prepare(*input, *output, *summaryPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	summary, err := preprocessing.Run(*input, *output, *summaryPath, *workers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Limpieza completada en %.2fs con %d workers\n", summary.ElapsedSeconds, summary.Workers)
	fmt.Printf("  Filas leidas:      %d\n", summary.TotalRead)
	fmt.Printf("  Filas validas:     %d\n", summary.Valid)
	fmt.Printf("  Filas descartadas: %d\n", summary.Discarded)
	for reason, count := range summary.DiscardedByReason {
		fmt.Printf("    - %s: %d\n", reason, count)
	}
	fmt.Printf("CSV limpio   -> %s\n", *output)
	fmt.Printf("Resumen JSON -> %s\n", *summaryPath)
}

// prepare descarga el CSV crudo si falta y crea las carpetas de salida, que
// no existen en un clon nuevo del repo (git no guarda carpetas vacias).
func prepare(input, output, summaryPath string) error {
	if err := dataset.EnsureRaw(input, os.Stdout); err != nil {
		return err
	}
	for _, path := range []string{output, summaryPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	return nil
}
