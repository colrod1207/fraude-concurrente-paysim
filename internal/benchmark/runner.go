package benchmark

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Result resume las mediciones de una configuracion (secuencial o N workers).
type Result struct {
	Task       string          // "limpieza" o "random-forest"
	Mode       string          // "secuencial" o "concurrente"
	Workers    int             // 1 para la version secuencial
	Runs       []time.Duration // tiempo de cada corrida cronometrada
	Mean       time.Duration   // media recortada de Runs
	Speedup    float64         // T-Secuencial / Mean
	Efficiency float64         // Speedup / Workers
	Profile    Profile         // corrida extra, separada del cronometro
}

// Profile describe el uso de recursos en una corrida.
type Profile struct {
	Wall         time.Duration
	CPU          time.Duration // tiempo de CPU del proceso (usuario + sistema)
	CoresUsed    float64       // CPU / Wall: nucleos ocupados en promedio
	CPUPercent   float64       // CoresUsed / runtime.NumCPU() * 100
	AllocMB      float64       // memoria total pedida al heap durante la corrida
	PeakHeapMB   float64       // maximo de heap en uso observado
	NumGoroutine int           // goroutines vivas al pico observado
}

// Time ejecuta fn runs veces y devuelve la duracion de cada corrida. Antes
// de cada una fuerza un GC para que la basura de la corrida anterior no se
// cobre en la siguiente.
func Time(runs int, fn func() error) ([]time.Duration, error) {
	if runs < 1 {
		return nil, fmt.Errorf("se necesita al menos una corrida")
	}
	durations := make([]time.Duration, runs)
	for i := range durations {
		runtime.GC()
		start := time.Now()
		if err := fn(); err != nil {
			return nil, err
		}
		durations[i] = time.Since(start)
	}
	return durations, nil
}

// sampleInterval es cada cuanto se lee la memoria durante ProfileRun.
const sampleInterval = 10 * time.Millisecond

// ProfileRun ejecuta fn una vez midiendo CPU y memoria. Leer MemStats
// detiene brevemente el programa, por eso esta corrida no se usa para
// calcular tiempos ni Speedup.
func ProfileRun(fn func() error) (Profile, error) {
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	cpuBefore, err := processCPUTime()
	if err != nil {
		return Profile{}, err
	}

	var peakHeap uint64
	var peakGoroutines int
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()
		var stats runtime.MemStats
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				runtime.ReadMemStats(&stats)
				if stats.HeapInuse > peakHeap {
					peakHeap = stats.HeapInuse
				}
				peakGoroutines = max(peakGoroutines, runtime.NumGoroutine())
			}
		}
	}()

	start := time.Now()
	runErr := fn()
	wall := time.Since(start)
	close(stop)
	wg.Wait() // despues de esto el muestreador ya no toca peakHeap
	if runErr != nil {
		return Profile{}, runErr
	}

	cpuAfter, err := processCPUTime()
	if err != nil {
		return Profile{}, err
	}
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	peakHeap = max(peakHeap, after.HeapInuse)

	cpu := cpuAfter - cpuBefore
	p := Profile{
		Wall:         wall,
		CPU:          cpu,
		AllocMB:      float64(after.TotalAlloc-before.TotalAlloc) / (1 << 20),
		PeakHeapMB:   float64(peakHeap) / (1 << 20),
		NumGoroutine: peakGoroutines,
	}
	if wall > 0 {
		p.CoresUsed = float64(cpu) / float64(wall)
		p.CPUPercent = p.CoresUsed / float64(runtime.NumCPU()) * 100
	}
	return p, nil
}
