// Ejecuta la limpieza concurrente del dataset PaySim (PC1, CC65).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/preprocessing"
)

func main() {
	input := flag.String("input", "data/raw/PS_20174392719_1491204439457_log.csv", "ruta del CSV crudo de PaySim")
	output := flag.String("output", "data/processed/paysim_clean.csv", "ruta del CSV limpio de salida")
	summaryPath := flag.String("summary", "data/processed/resumen_limpieza.json", "ruta del resumen JSON")
	workers := flag.Int("workers", 4, "numero de goroutines worker")
	flag.Parse()

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
