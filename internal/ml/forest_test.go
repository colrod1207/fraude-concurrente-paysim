package ml

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestBootstrap_SizeAndReplacement(t *testing.T) {
	samples := []Sample{
		{[]float64{1}, 0}, {[]float64{2}, 0},
		{[]float64{3}, 1}, {[]float64{4}, 1},
	}
	bag := bootstrap(samples, rand.New(rand.NewSource(1)))
	if len(bag) != len(samples) {
		t.Fatalf("bootstrap size = %d, want %d", len(bag), len(samples))
	}
	seen := make(map[float64]bool)
	duplicate := false
	for _, sample := range bag {
		value := sample.Features[0]
		found := false
		for _, original := range samples {
			if reflect.DeepEqual(sample, original) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("bootstrap contains an unknown sample: %v", sample)
		}
		duplicate = duplicate || seen[value]
		seen[value] = true
	}
	if !duplicate {
		t.Fatal("expected repeated samples with this fixed seed")
	}
}

func TestForest_CountDeterminismAndExecutionOrder(t *testing.T) {
	samples := []Sample{
		{[]float64{1, 8, 0, 2}, 0}, {[]float64{2, 1, 1, 3}, 0},
		{[]float64{3, 6, 0, 1}, 1}, {[]float64{4, 2, 1, 4}, 1},
		{[]float64{5, 7, 1, 0}, 0}, {[]float64{6, 3, 0, 5}, 1},
	}
	const seed int64 = 42
	forest, err := TrainForest(samples, 7, 3, seed)
	if err != nil {
		t.Fatal(err)
	}
	if len(forest.trees) != 7 || forest.numFeatures != 4 {
		t.Fatalf("unexpected forest size or feature count: %+v", forest)
	}
	again, err := TrainForest(samples, 7, 3, seed)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(forest, again) {
		t.Fatal("the same seed and samples must reproduce every tree")
	}
	// Reentrenar en orden inverso sigue siendo secuencial y comprueba que
	// ningun arbol depende del estado aleatorio consumido por otro arbol.
	master := rand.New(rand.NewSource(seed))
	seeds := make([]int64, 7)
	for i := range seeds {
		seeds[i] = master.Int63()
	}
	for i := len(seeds) - 1; i >= 0; i-- {
		tree := trainForestTree(samples, 3, 2, seeds[i])
		if !reflect.DeepEqual(tree, forest.trees[i]) {
			t.Fatalf("tree %d changed when trained in reverse order", i)
		}
	}
}

func TestCandidateFeatures(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for _, count := range []int{1, 2, 4, 10} {
		mtry := max(1, int(math.Sqrt(float64(count))))
		features := candidateFeatures(count, mtry, rng)
		if len(features) != mtry {
			t.Fatalf("%d features: got %d candidates, want %d", count, len(features), mtry)
		}
		seen := make(map[int]bool)
		for _, feature := range features {
			if feature < 0 || feature >= count || seen[feature] {
				t.Fatalf("invalid or repeated candidate: %v", features)
			}
			seen[feature] = true
		}
	}
	if got := candidateFeatures(4, 4, nil); !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
		t.Fatalf("without RNG all features must keep their original order: %v", got)
	}
}

func TestTree_SelectsFeaturesAtEachNode(t *testing.T) {
	// Las cuatro features son iguales para que cualquiera permita dividir.
	samples := []Sample{
		{[]float64{1, 1, 1, 1}, 0}, {[]float64{2, 2, 2, 2}, 0},
		{[]float64{3, 3, 3, 3}, 1}, {[]float64{4, 4, 4, 4}, 1},
		{[]float64{5, 5, 5, 5}, 0}, {[]float64{6, 6, 6, 6}, 0},
	}
	root := buildTree(samples, 0, 2, 1, rand.New(rand.NewSource(42)))
	expected := rand.New(rand.NewSource(42))
	rootFeature := expected.Perm(4)[0]
	rightFeature := expected.Perm(4)[0]
	if root.left == nil || root.right == nil || root.right.left == nil {
		t.Fatal("expected a root split and another split in its right child")
	}
	if root.feature != rootFeature || root.right.feature != rightFeature {
		t.Fatalf("expected a fresh selection per node: root=%d, right=%d", root.feature, root.right.feature)
	}
}

func TestForest_MajorityVote(t *testing.T) {
	for _, tc := range []struct {
		name   string
		labels []int
		want   int
	}{
		{"majority one", []int{0, 1, 1}, 1},
		{"majority zero", []int{1, 0, 0}, 0},
		{"tie", []int{1, 0}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			forest := &RandomForest{numFeatures: 1}
			for _, label := range tc.labels {
				// Arboles de una sola hoja para controlar exactamente los votos.
				forest.trees = append(forest.trees, &DecisionTree{
					root: &node{label: label}, numFeatures: 1,
				})
			}
			got, err := forest.Predict([]float64{5})
			if err != nil || got != tc.want {
				t.Fatalf("prediction = %d, %v; want %d", got, err, tc.want)
			}
		})
	}
}

func TestForest_LearnsSimpleDataset(t *testing.T) {
	samples := []Sample{
		{[]float64{0}, 0}, {[]float64{1}, 0}, {[]float64{2}, 0},
		{[]float64{8}, 1}, {[]float64{9}, 1}, {[]float64{10}, 1},
	}
	forest, err := TrainForest(samples, 15, 2, 42)
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []Sample{{[]float64{1.5}, 0}, {[]float64{8.5}, 1}} {
		got, err := forest.Predict(sample.Features)
		if err != nil || got != sample.Label {
			t.Errorf("Predict(%v) = %d, %v; want %d", sample.Features, got, err, sample.Label)
		}
	}
}

func TestTrainForest_InvalidInput(t *testing.T) {
	valid := []Sample{{[]float64{1}, 0}}
	for _, tc := range []struct {
		name               string
		samples            []Sample
		numTrees, maxDepth int
	}{
		{"zero trees", valid, 0, 1},
		{"negative trees", valid, -1, 1},
		{"empty", nil, 1, 1},
		{"no features", []Sample{{nil, 0}}, 1, 1},
		{"invalid label", []Sample{{[]float64{1}, 2}}, 1, 1},
		{"different lengths", []Sample{{[]float64{1}, 0}, {[]float64{1, 2}, 1}}, 1, 1},
		{"NaN", []Sample{{[]float64{math.NaN()}, 0}}, 1, 1},
		{"infinity", []Sample{{[]float64{math.Inf(1)}, 0}}, 1, 1},
		{"negative depth", valid, 1, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := TrainForest(tc.samples, tc.numTrees, tc.maxDepth, 42); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestForestPredict_InvalidInput(t *testing.T) {
	forest, err := TrainForest([]Sample{{[]float64{1}, 0}}, 1, 0, 42)
	if err != nil {
		t.Fatal(err)
	}
	for _, features := range [][]float64{nil, {1, 2}, {math.NaN()}, {math.Inf(-1)}} {
		if _, err := forest.Predict(features); err == nil {
			t.Errorf("expected an error for features %v", features)
		}
	}
	for _, untrained := range []*RandomForest{nil, {}} {
		if _, err := untrained.Predict([]float64{1}); err == nil {
			t.Fatal("expected an error for an untrained forest")
		}
	}
}
