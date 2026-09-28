// Package benchmark mide tiempos de las versiones secuencial y concurrente
// y calcula la media recortada, el Speedup y la eficiencia.
package benchmark

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// TrimmedMean ordena las corridas, descarta floor(n*trim) de cada extremo
// (las mas rapidas y las mas lentas) y promedia el resto. Asi un pico
// aislado (otro programa usando la CPU, el disco en cache, etc.) no
// distorsiona el tiempo reportado. trim debe estar en [0, 0.5).
func TrimmedMean(runs []time.Duration, trim float64) (time.Duration, error) {
	if len(runs) == 0 {
		return 0, fmt.Errorf("se necesita al menos una corrida")
	}
	if math.IsNaN(trim) || trim < 0 || trim >= 0.5 {
		return 0, fmt.Errorf("el recorte debe estar en [0, 0.5)")
	}
	sorted := append([]time.Duration(nil), runs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	cut := int(float64(len(sorted)) * trim)
	kept := sorted[cut : len(sorted)-cut]
	var total time.Duration
	for _, run := range kept {
		total += run
	}
	return total / time.Duration(len(kept)), nil
}

// Speedup = T-Secuencial / T-Concurrente. Devuelve 0 si T-Concurrente es 0.
func Speedup(sequential, concurrent time.Duration) float64 {
	if concurrent == 0 {
		return 0
	}
	return float64(sequential) / float64(concurrent)
}

// Efficiency = Speedup / workers: que fraccion del ideal lineal se alcanza.
func Efficiency(speedup float64, workers int) float64 {
	if workers == 0 {
		return 0
	}
	return speedup / float64(workers)
}
