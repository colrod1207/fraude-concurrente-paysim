package ml

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

type node struct {
	feature   int
	threshold float64
	label     int
	left      *node
	right     *node
}

// DecisionTree clasifica mediante comparaciones feature <= threshold.
// Un nodo sin hijos es una hoja y devuelve su clase mayoritaria.
type DecisionTree struct {
	root        *node
	numFeatures int
}

// TrainTree construye el arbol de forma secuencial, sin modificar samples.
// La raiz tiene profundidad 0; maxDepth=0 produce una unica hoja.
func TrainTree(samples []Sample, maxDepth int) (*DecisionTree, error) {
	numFeatures, err := validateTraining(samples, maxDepth)
	if err != nil {
		return nil, err
	}
	return &DecisionTree{
		root:        buildTree(samples, 0, maxDepth, numFeatures, nil),
		numFeatures: numFeatures,
	}, nil
}

// validateTraining se comparte con el bosque para validar antes del bootstrap.
func validateTraining(samples []Sample, maxDepth int) (int, error) {
	if len(samples) == 0 {
		return 0, fmt.Errorf("se necesita al menos una muestra")
	}
	if maxDepth < 0 {
		return 0, fmt.Errorf("la profundidad maxima no puede ser negativa")
	}
	numFeatures := len(samples[0].Features)
	if numFeatures == 0 {
		return 0, fmt.Errorf("se necesita al menos una feature")
	}
	for i, sample := range samples {
		if sample.Label != 0 && sample.Label != 1 {
			return 0, fmt.Errorf("muestra %d: label debe ser 0 o 1", i)
		}
		if err := validateFeatures(sample.Features, numFeatures); err != nil {
			return 0, fmt.Errorf("muestra %d: %w", i, err)
		}
	}
	return numFeatures, nil
}

// Predict recorre un solo camino desde la raiz hasta una hoja.
func (t *DecisionTree) Predict(features []float64) (int, error) {
	if t == nil || t.root == nil {
		return 0, fmt.Errorf("el arbol no esta entrenado")
	}
	if err := validateFeatures(features, t.numFeatures); err != nil {
		return 0, err
	}
	n := t.root
	for n.left != nil {
		if features[n.feature] <= n.threshold {
			n = n.left
		} else {
			n = n.right
		}
	}
	return n.label, nil
}

func validateFeatures(features []float64, expected int) error {
	if len(features) != expected {
		return fmt.Errorf("se esperaban %d features, se recibieron %d", expected, len(features))
	}
	for _, value := range features {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("las features deben ser numeros finitos")
		}
	}
	return nil
}

// gini mide la mezcla de clases: 0 para un nodo puro, 0.5 para mitad y mitad.
func gini(zeros, ones int) float64 {
	total := zeros + ones
	if total == 0 {
		return 0
	}
	p0 := float64(zeros) / float64(total)
	p1 := float64(ones) / float64(total)
	return 1 - p0*p0 - p1*p1
}

// Sin RNG se evaluan todas las features; con RNG se eligen mtry por nodo.
func candidateFeatures(numFeatures, mtry int, rng *rand.Rand) []int {
	if rng != nil {
		return rng.Perm(numFeatures)[:mtry]
	}
	features := make([]int, numFeatures)
	for i := range features {
		features[i] = i
	}
	return features
}

func buildTree(samples []Sample, depth, maxDepth, mtry int, rng *rand.Rand) *node {
	ones := 0
	for _, sample := range samples {
		ones += sample.Label
	}
	zeros := len(samples) - ones
	n := &node{}
	if ones > zeros { // En empate se conserva la clase 0.
		n.label = 1
	}
	if len(samples) < 2 || zeros == 0 || ones == 0 || depth >= maxDepth {
		return n
	}

	bestGini := gini(zeros, ones)
	bestFeature := -1
	var bestThreshold float64
	// Se ordena una copia para no cambiar el orden de las muestras del usuario.
	ordered := append([]Sample(nil), samples...)
	for _, feature := range candidateFeatures(len(samples[0].Features), mtry, rng) {
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].Features[feature] < ordered[j].Features[feature]
		})
		leftOnes := 0
		for i := 0; i < len(ordered)-1; i++ {
			leftOnes += ordered[i].Label
			lower := ordered[i].Features[feature]
			upper := ordered[i+1].Features[feature]
			if lower == upper {
				continue // Valores iguales no pueden separarse con un umbral.
			}
			leftCount := i + 1
			rightCount := len(ordered) - leftCount
			rightOnes := ones - leftOnes
			weighted := (float64(leftCount)*gini(leftCount-leftOnes, leftOnes) +
				float64(rightCount)*gini(rightCount-rightOnes, rightOnes)) / float64(len(ordered))
			if weighted < bestGini {
				bestGini = weighted
				bestFeature = feature
				bestThreshold = lower/2 + upper/2
				// El redondeo no debe llevar el umbral al siguiente valor.
				if bestThreshold < lower || bestThreshold >= upper {
					bestThreshold = lower
				}
			}
		}
	}
	if bestFeature == -1 {
		return n // Ninguna division reduce la impureza del nodo.
	}

	var left, right []Sample
	for _, sample := range samples {
		if sample.Features[bestFeature] <= bestThreshold {
			left = append(left, sample)
		} else {
			right = append(right, sample)
		}
	}
	n.feature = bestFeature
	n.threshold = bestThreshold
	n.left = buildTree(left, depth+1, maxDepth, mtry, rng)
	n.right = buildTree(right, depth+1, maxDepth, mtry, rng)
	return n
}
