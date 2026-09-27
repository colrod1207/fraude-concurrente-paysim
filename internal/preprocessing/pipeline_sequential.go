package preprocessing

import (
	"encoding/csv"
	"os"
	"time"
)

// RunSequential procesa el dataset sin concurrencia: una sola goroutine
// lee cada fila y llama a ProcessRow directamente, sin worker pool ni
// channel de resultados. Sirve como baseline (T-Secuencial) para calcular
// el Speedup de Run(numWorkers) = T-Secuencial / T-Concurrente.
//
// Reutiliza StreamCSV para que la lectura del archivo sea idéntica a la
// versión concurrente (misma lógica de parseo de filas); la única
// diferencia real entre esta función y Run() es que aquí no hay pool de
// workers: cada fila se procesa en el mismo goroutine que la reductora.
func RunSequential(inputPath, outputPath, summaryPath string) (*Summary, error) {
	start := time.Now()

	rows := make(chan []string, channelBuffer)
	readErrCh := make(chan error, 1)
	go func() {
		readErrCh <- StreamCSV(inputPath, rows)
	}()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	if err := writer.Write(Header()); err != nil {
		return nil, err
	}

	summary := NewSummary()
	// Sin worker pool: cada fila se valida y transforma aquí mismo, una a
	// la vez, antes de pasar a la siguiente. No hay paralelismo posible.
	for fields := range rows {
		result := ProcessRow(fields)
		summary.TotalRead++
		if result.Discard != "" {
			summary.Discarded++
			summary.DiscardedByReason[result.Discard]++
			continue
		}
		summary.Valid++
		if err := writer.Write(recordToRow(result.Record)); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	if err := <-readErrCh; err != nil {
		return nil, err
	}

	summary.Workers = 1 // referencial: no hay pool, se deja en 1 para el JSON
	summary.ElapsedSeconds = time.Since(start).Seconds()
	if err := summary.WriteJSON(summaryPath); err != nil {
		return nil, err
	}
	return summary, nil
}
