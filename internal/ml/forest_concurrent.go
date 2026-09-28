package ml

import (
	"fmt"
	"sync"
)

// TrainForestConcurrent entrena los mismos arboles que TrainForest usando un
// pool de numWorkers goroutines. Con la misma semilla produce exactamente el
// mismo bosque que la version secuencial, para cualquier numero de workers.
//
// Patron: los indices de arbol se cargan en un channel con buffer que luego
// se cierra; cada worker toma un indice, entrena ese arbol con su RNG exclusivo y lo
// guarda en forest.trees[i]. Cada indice lo recibe un solo worker, asi que
// ningun par de goroutines escribe la misma posicion y no hace falta ningun
// lock. Las muestras se comparten solo para lectura. El sync.WaitGroup marca
// cuando todos los arboles estan listos.
func TrainForestConcurrent(samples []Sample, numTrees, maxDepth int, seed int64, numWorkers int) (*RandomForest, error) {
	if numWorkers < 1 {
		return nil, fmt.Errorf("el numero de workers debe ser positivo")
	}
	forest, mtry, seeds, err := prepareForest(samples, numTrees, maxDepth, seed)
	if err != nil {
		return nil, err
	}

	jobs := make(chan int, numTrees)
	for i := range seeds {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wg.Done()
			for i := range jobs {
				forest.trees[i] = trainForestTree(samples, maxDepth, mtry, seeds[i])
			}
		}()
	}
	wg.Wait()
	return forest, nil
}
