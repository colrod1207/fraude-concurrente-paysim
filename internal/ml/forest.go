package ml

import (
	"fmt"
	"math"
	"math/rand"
)

// RandomForest combina por votacion arboles entrenados secuencialmente.
type RandomForest struct {
	trees       []*DecisionTree
	numFeatures int
}

// TrainForest entrena numTrees arboles, terminando uno antes de iniciar otro.
// La misma semilla, muestras en el mismo orden y parametros reproducen el bosque.
func TrainForest(samples []Sample, numTrees, maxDepth int, seed int64) (*RandomForest, error) {
	if numTrees < 1 {
		return nil, fmt.Errorf("el numero de arboles debe ser positivo")
	}
	numFeatures, err := validateTraining(samples, maxDepth)
	if err != nil {
		return nil, err
	}
	mtry := max(1, int(math.Sqrt(float64(numFeatures))))

	// Se asigna cada semilla a su indice antes de entrenar. Un futuro ejecutor
	// concurrente podra conservar esta asignacion sin depender del orden de ejecucion.
	master := rand.New(rand.NewSource(seed))
	seeds := make([]int64, numTrees)
	for i := range seeds {
		seeds[i] = master.Int63()
	}
	forest := &RandomForest{
		trees:       make([]*DecisionTree, numTrees),
		numFeatures: numFeatures,
	}
	for i, treeSeed := range seeds {
		forest.trees[i] = trainForestTree(samples, maxDepth, mtry, treeSeed)
	}
	return forest, nil
}

// trainForestTree recibe datos ya validados y usa un RNG exclusivo del arbol.
func trainForestTree(samples []Sample, maxDepth, mtry int, seed int64) *DecisionTree {
	rng := rand.New(rand.NewSource(seed))
	bag := bootstrap(samples, rng)
	return &DecisionTree{
		root:        buildTree(bag, 0, maxDepth, mtry, rng),
		numFeatures: len(samples[0].Features),
	}
}

// bootstrap extrae N muestras con reemplazo de las N muestras originales.
// Las features se comparten solo para lectura; el entrenamiento no las modifica.
func bootstrap(samples []Sample, rng *rand.Rand) []Sample {
	bag := make([]Sample, len(samples))
	for i := range bag {
		bag[i] = samples[rng.Intn(len(samples))]
	}
	return bag
}

// Predict cuenta los votos de todos los arboles. En empate devuelve clase 0.
func (f *RandomForest) Predict(features []float64) (int, error) {
	if f == nil || len(f.trees) == 0 {
		return 0, fmt.Errorf("el bosque no esta entrenado")
	}
	if err := validateFeatures(features, f.numFeatures); err != nil {
		return 0, err
	}
	ones := 0
	for _, tree := range f.trees {
		label, err := tree.Predict(features)
		if err != nil {
			return 0, err
		}
		ones += label
	}
	if ones > len(f.trees)-ones {
		return 1, nil
	}
	return 0, nil
}
