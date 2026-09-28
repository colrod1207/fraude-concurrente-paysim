// Ejecuta el Random Forest (secuencial o concurrente) sobre el CSV limpio de PaySim.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/dataset"
	"github.com/colrod1207/fraude-concurrente-paysim/internal/ml"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	input := flag.String("input", dataset.DefaultClean, "ruta del CSV limpio (se genera si falta)")
	trees := flag.Int("trees", 10, "numero de arboles")
	depth := flag.Int("depth", 5, "profundidad maxima (raiz = 0)")
	seed := flag.Int64("seed", 42, "semilla para split y entrenamiento, con generadores separados")
	limit := flag.Int("limit", 10000, "primeros N registros a cargar; 0 lee todos (puede ser costoso)")
	workers := flag.Int("workers", 0, "goroutines para entrenar; 0 usa la version secuencial")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("use los parametros --input, --trees, --depth, --seed, --limit y --workers")
	}
	if *trees < 1 || *depth < 0 || *limit < 0 || *workers < 0 {
		return fmt.Errorf("trees debe ser positivo; depth, limit y workers no pueden ser negativos")
	}
	if err := dataset.EnsureClean(dataset.DefaultRaw, *input, dataset.DefaultSummary, runtime.NumCPU(), os.Stdout); err != nil {
		return err
	}
	samples, err := ml.LoadCleanCSV(*input, *limit)
	if err != nil {
		return err
	}
	train, test, err := ml.StratifiedSplit(samples, *seed)
	if err != nil {
		return err
	}
	fmt.Printf("Samples: %d\nTrain: %d\nTest: %d\n", len(samples), len(train), len(test))
	fmt.Printf("Trees: %d\nMax depth: %d\nSeed: %d\nLimit: %d\n", *trees, *depth, *seed, *limit)
	fmt.Println("Split: 80/20 estratificado (decision tecnica del proyecto)")
	if *workers == 0 {
		fmt.Println("Mode: secuencial")
	} else {
		fmt.Printf("Mode: concurrente (%d workers)\n", *workers)
	}

	// Solo mide una ejecucion del entrenamiento; no es el benchmark oficial.
	start := time.Now()
	var forest *ml.RandomForest
	if *workers == 0 {
		forest, err = ml.TrainForest(train, *trees, *depth, *seed)
	} else {
		forest, err = ml.TrainForestConcurrent(train, *trees, *depth, *seed, *workers)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Training time: %s\n", time.Since(start))
	actual := make([]int, len(test))
	predicted := make([]int, len(test))
	for i, sample := range test {
		actual[i] = sample.Label
		predicted[i], err = forest.Predict(sample.Features)
		if err != nil {
			return err
		}
	}
	metrics, err := ml.CalculateMetrics(actual, predicted)
	if err != nil {
		return err
	}
	fmt.Printf("\nConfusion matrix:\nTP: %d\nTN: %d\nFP: %d\nFN: %d\n",
		metrics.TP, metrics.TN, metrics.FP, metrics.FN)
	fmt.Printf("\nAccuracy: %.6f\nPrecision: %.6f\nRecall: %.6f\nF1: %.6f\n",
		metrics.Accuracy, metrics.Precision, metrics.Recall, metrics.F1)
	return nil
}
