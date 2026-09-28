package ml

import (
	"math/rand"
	"reflect"
	"testing"
)

// randomSamples genera un dataset sintetico reproducible para comparar
// la version secuencial con la concurrente sobre datos no triviales.
func randomSamples(n, numFeatures int, seed int64) []Sample {
	rng := rand.New(rand.NewSource(seed))
	samples := make([]Sample, n)
	for i := range samples {
		features := make([]float64, numFeatures)
		for j := range features {
			features[j] = rng.Float64() * 100
		}
		label := 0
		if features[0]+features[1] > 110 {
			label = 1
		}
		samples[i] = Sample{Features: features, Label: label}
	}
	return samples
}

func TestTrainForestConcurrent_SameForestAsSequential(t *testing.T) {
	samples := randomSamples(300, 6, 7)
	const numTrees, maxDepth, seed = 13, 4, 42
	sequential, err := TrainForest(samples, numTrees, maxDepth, seed)
	if err != nil {
		t.Fatal(err)
	}
	// Incluye mas workers que arboles para cubrir workers ociosos.
	for _, workers := range []int{1, 2, 4, 8, 32} {
		concurrent, err := TrainForestConcurrent(samples, numTrees, maxDepth, seed, workers)
		if err != nil {
			t.Fatalf("workers=%d: %v", workers, err)
		}
		if !reflect.DeepEqual(sequential, concurrent) {
			t.Fatalf("workers=%d: el bosque concurrente difiere del secuencial", workers)
		}
	}
}

func TestTrainForestConcurrent_DoesNotModifySamples(t *testing.T) {
	samples := randomSamples(100, 4, 3)
	original := make([]Sample, len(samples))
	for i, sample := range samples {
		original[i] = Sample{append([]float64(nil), sample.Features...), sample.Label}
	}
	if _, err := TrainForestConcurrent(samples, 8, 3, 1, 4); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(samples, original) {
		t.Fatal("el entrenamiento concurrente no debe modificar las muestras")
	}
}

func TestTrainForestConcurrent_InvalidInput(t *testing.T) {
	valid := []Sample{{[]float64{1}, 0}}
	for _, tc := range []struct {
		name                        string
		samples                     []Sample
		numTrees, maxDepth, workers int
	}{
		{"zero workers", valid, 1, 1, 0},
		{"negative workers", valid, 1, 1, -2},
		{"zero trees", valid, 0, 1, 2},
		{"empty", nil, 1, 1, 2},
		{"negative depth", valid, 1, -1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := TrainForestConcurrent(tc.samples, tc.numTrees, tc.maxDepth, 42, tc.workers); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
