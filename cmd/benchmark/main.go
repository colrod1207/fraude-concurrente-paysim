// Cronometra la version secuencial y la concurrente de la limpieza y del
// Random Forest, y calcula media recortada, Speedup, eficiencia, CPU y memoria.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/benchmark"
	"github.com/colrod1207/fraude-concurrente-paysim/internal/ml"
	"github.com/colrod1207/fraude-concurrente-paysim/internal/preprocessing"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	task := flag.String("task", "all", "que medir: all, limpieza o forest")
	raw := flag.String("raw", "data/raw/PS_20174392719_1491204439457_log.csv", "CSV crudo (para la limpieza)")
	clean := flag.String("clean", "data/processed/paysim_clean.csv", "CSV limpio (para el Random Forest)")
	workersFlag := flag.String("workers", "1,2,4,8", "numeros de workers a probar, separados por coma")
	runs := flag.Int("runs", 10, "corridas cronometradas por configuracion")
	trim := flag.Float64("trim", 0.1, "fraccion descartada en cada extremo para la media recortada")
	trees := flag.Int("trees", 32, "arboles del Random Forest")
	depth := flag.Int("depth", 8, "profundidad maxima de cada arbol")
	limit := flag.Int("limit", 200000, "filas del CSV limpio a cargar para el Random Forest; 0 lee todas")
	seed := flag.Int64("seed", 42, "semilla del split y del bosque")
	out := flag.String("out", "results", "carpeta donde se guardan benchmark.csv y benchmark.md")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("argumento inesperado: %s", flag.Arg(0))
	}

	workers, err := parseWorkers(*workersFlag)
	if err != nil {
		return err
	}
	cfg := benchmark.Config{Runs: *runs, Trim: *trim, Workers: workers}
	if *runs < 1 || !(*trim >= 0 && *trim < 0.5) {
		return fmt.Errorf("runs debe ser positivo y trim debe estar en [0, 0.5)")
	}

	fmt.Printf("Equipo: %s/%s, %d CPUs logicas, GOMAXPROCS=%d, %s\n",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.GOMAXPROCS(0), runtime.Version())
	fmt.Printf("Corridas: %d por configuracion, media recortada al %.0f%% por extremo\n\n", *runs, *trim*100)

	var results []benchmark.Result
	switch *task {
	case "all", "limpieza", "forest":
	default:
		return fmt.Errorf("task debe ser all, limpieza o forest")
	}
	if *task == "all" || *task == "limpieza" {
		r, err := benchmarkCleaning(*raw, cfg)
		if err != nil {
			return err
		}
		results = append(results, r...)
	}
	if *task == "all" || *task == "forest" {
		r, err := benchmarkForest(*clean, *limit, *trees, *depth, *seed, cfg)
		if err != nil {
			return err
		}
		results = append(results, r...)
	}

	if err := benchmark.WriteMarkdown(os.Stdout, results); err != nil {
		return err
	}
	return saveResults(*out, results)
}

func benchmarkCleaning(raw string, cfg benchmark.Config) ([]benchmark.Result, error) {
	if _, err := os.Stat(raw); err != nil {
		return nil, fmt.Errorf("no se encuentra el CSV crudo (ver README para descargarlo): %w", err)
	}
	// Las salidas de cada corrida se descartan: solo interesa el tiempo.
	tmp, err := os.MkdirTemp("", "benchmark-limpieza-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	output := filepath.Join(tmp, "clean.csv")
	summary := filepath.Join(tmp, "resumen.json")

	fmt.Println("Midiendo limpieza...")
	return benchmark.Compare("limpieza", cfg,
		func() error {
			_, err := preprocessing.RunSequential(raw, output, summary)
			return err
		},
		func(workers int) error {
			_, err := preprocessing.Run(raw, output, summary, workers)
			return err
		})
}

func benchmarkForest(clean string, limit, trees, depth int, seed int64, cfg benchmark.Config) ([]benchmark.Result, error) {
	// La carga y el split se hacen una sola vez: solo se cronometra el entrenamiento.
	samples, err := ml.LoadCleanCSV(clean, limit)
	if err != nil {
		return nil, fmt.Errorf("cargar CSV limpio (correr antes la limpieza): %w", err)
	}
	train, _, err := ml.StratifiedSplit(samples, seed)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Midiendo Random Forest (%d muestras de entrenamiento, %d arboles, profundidad %d)...\n",
		len(train), trees, depth)
	return benchmark.Compare("random-forest", cfg,
		func() error {
			_, err := ml.TrainForest(train, trees, depth, seed)
			return err
		},
		func(workers int) error {
			_, err := ml.TrainForestConcurrent(train, trees, depth, seed, workers)
			return err
		})
}

func parseWorkers(value string) ([]int, error) {
	var workers []int
	for _, part := range strings.Split(value, ",") {
		w, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || w < 1 {
			return nil, fmt.Errorf("workers invalido %q: use enteros positivos separados por coma", part)
		}
		workers = append(workers, w)
	}
	return workers, nil
}

func saveResults(dir string, results []benchmark.Result) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, write := range map[string]func(*os.File) error{
		"benchmark.csv": func(f *os.File) error { return benchmark.WriteCSV(f, results) },
		"benchmark.md":  func(f *os.File) error { return benchmark.WriteMarkdown(f, results) },
	} {
		path := filepath.Join(dir, name)
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := write(f); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		fmt.Printf("\nResultados -> %s", path)
	}
	fmt.Println()
	return nil
}
