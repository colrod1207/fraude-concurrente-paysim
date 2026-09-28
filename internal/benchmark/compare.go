package benchmark

import "fmt"

// Config fija cuantas corridas se cronometran y cuanto se recorta.
type Config struct {
	Runs    int     // corridas cronometradas por configuracion
	Trim    float64 // fraccion descartada en cada extremo para la media recortada
	Workers []int   // numeros de workers a probar en la version concurrente
}

// Compare mide la version secuencial y luego la concurrente con cada numero
// de workers. El Speedup de cada fila se calcula contra la media recortada
// de la secuencial, que se reporta con Speedup 1.
func Compare(task string, cfg Config, sequential func() error, concurrent func(workers int) error) ([]Result, error) {
	if len(cfg.Workers) == 0 {
		return nil, fmt.Errorf("se necesita al menos un numero de workers")
	}
	for _, w := range cfg.Workers {
		if w < 1 {
			return nil, fmt.Errorf("numero de workers invalido: %d", w)
		}
	}
	base, err := measure(task, "secuencial", 1, cfg, sequential)
	if err != nil {
		return nil, fmt.Errorf("%s secuencial: %w", task, err)
	}
	base.Speedup, base.Efficiency = 1, 1
	results := []Result{base}
	for _, w := range cfg.Workers {
		r, err := measure(task, "concurrente", w, cfg, func() error { return concurrent(w) })
		if err != nil {
			return nil, fmt.Errorf("%s concurrente con %d workers: %w", task, w, err)
		}
		r.Speedup = Speedup(base.Mean, r.Mean)
		r.Efficiency = Efficiency(r.Speedup, w)
		results = append(results, r)
	}
	return results, nil
}

func measure(task, mode string, workers int, cfg Config, fn func() error) (Result, error) {
	runs, err := Time(cfg.Runs, fn)
	if err != nil {
		return Result{}, err
	}
	mean, err := TrimmedMean(runs, cfg.Trim)
	if err != nil {
		return Result{}, err
	}
	profile, err := ProfileRun(fn)
	if err != nil {
		return Result{}, err
	}
	return Result{Task: task, Mode: mode, Workers: workers, Runs: runs, Mean: mean, Profile: profile}, nil
}
