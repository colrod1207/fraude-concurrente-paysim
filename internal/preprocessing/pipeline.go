package preprocessing

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// channelBuffer amortigua la diferencia de velocidad entre el productor y
// el pool de workers sin que el diseño deje de ser productor/consumidor.
const channelBuffer = 256

// Run orquesta el pipeline completo: una goroutine productora
// (StreamCSV) alimenta un channel de filas crudas, un pool de numWorkers
// goroutines las consume y valida en paralelo (ProcessRow), y la propia
// goroutine que llama a Run actúa como único reductor: escribe el CSV
// limpio y acumula el resumen. Al ser el único lugar que toca ese estado,
// no hace falta ningún lock.
func Run(inputPath, outputPath, summaryPath string, numWorkers int) (*Summary, error) {
	start := time.Now()
	if numWorkers < 1 {
		numWorkers = 1
	}

	rows := make(chan []string, channelBuffer)
	results := make(chan Result, channelBuffer)

	readErrCh := make(chan error, 1)
	go func() {
		readErrCh <- StreamCSV(inputPath, rows)
	}()

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for fields := range rows {
				results <- ProcessRow(fields)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
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
	for result := range results {
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

	summary.Workers = numWorkers
	summary.ElapsedSeconds = time.Since(start).Seconds()
	if err := summary.WriteJSON(summaryPath); err != nil {
		return nil, err
	}
	return summary, nil
}

func recordToRow(r *CleanRecord) []string {
	return []string{
		strconv.Itoa(r.Step), r.Type, strconv.Itoa(r.TypeCode), formatAmount(r.Amount),
		r.NameOrig, formatAmount(r.OldBalanceOrg), formatAmount(r.NewBalanceOrig),
		r.NameDest, formatAmount(r.OldBalanceDest), formatAmount(r.NewBalanceDest),
		strconv.Itoa(r.IsFraud), strconv.Itoa(r.IsFlaggedFraud),
		formatAmount(r.ErrorBalanceOrig), formatAmount(r.ErrorBalanceDest),
		strconv.Itoa(r.IsMerchantDest), strconv.Itoa(r.Hour),
	}
}

func formatAmount(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
