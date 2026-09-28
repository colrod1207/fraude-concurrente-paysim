package benchmark

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var csvHeader = []string{
	"tarea", "modo", "workers", "corridas", "media_recortada_s", "speedup",
	"eficiencia", "cpu_porcentaje", "nucleos_usados", "heap_pico_mb",
	"memoria_asignada_mb", "tiempos_s",
}

// WriteCSV escribe una fila por configuracion, con todas las corridas al
// final para poder recalcular cualquier estadistica despues.
func WriteCSV(w io.Writer, results []Result) error {
	writer := csv.NewWriter(w)
	if err := writer.Write(csvHeader); err != nil {
		return err
	}
	for _, r := range results {
		runs := make([]string, len(r.Runs))
		for i, run := range r.Runs {
			runs[i] = strconv.FormatFloat(run.Seconds(), 'f', 4, 64)
		}
		row := []string{
			r.Task, r.Mode, strconv.Itoa(r.Workers), strconv.Itoa(len(r.Runs)),
			strconv.FormatFloat(r.Mean.Seconds(), 'f', 4, 64),
			strconv.FormatFloat(r.Speedup, 'f', 3, 64),
			strconv.FormatFloat(r.Efficiency, 'f', 3, 64),
			strconv.FormatFloat(r.Profile.CPUPercent, 'f', 1, 64),
			strconv.FormatFloat(r.Profile.CoresUsed, 'f', 2, 64),
			strconv.FormatFloat(r.Profile.PeakHeapMB, 'f', 1, 64),
			strconv.FormatFloat(r.Profile.AllocMB, 'f', 1, 64),
			strings.Join(runs, ";"),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

// WriteMarkdown escribe la tabla de resultados lista para pegar en el informe.
func WriteMarkdown(w io.Writer, results []Result) error {
	_, err := fmt.Fprintln(w, "| Tarea | Modo | Workers | Media recortada (s) | Speedup | Eficiencia | CPU (%) | Núcleos usados | Heap pico (MB) | Memoria asignada (MB) |\n"+
		"|---|---|---:|---:|---:|---:|---:|---:|---:|---:|")
	if err != nil {
		return err
	}
	for _, r := range results {
		_, err := fmt.Fprintf(w, "| %s | %s | %d | %.4f | %.2f | %.2f | %.1f | %.2f | %.1f | %.1f |\n",
			r.Task, r.Mode, r.Workers, r.Mean.Seconds(), r.Speedup, r.Efficiency,
			r.Profile.CPUPercent, r.Profile.CoresUsed, r.Profile.PeakHeapMB, r.Profile.AllocMB)
		if err != nil {
			return err
		}
	}
	return nil
}
