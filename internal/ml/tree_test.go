package ml

import (
	"math"
	"reflect"
	"testing"
)

func TestGini(t *testing.T) {
	for _, tc := range []struct {
		zeros, ones int
		want        float64
	}{
		{0, 0, 0},
		{4, 0, 0},
		{0, 4, 0},
		{2, 2, 0.5},
		{3, 1, 0.375},
	} {
		if got := gini(tc.zeros, tc.ones); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("gini(%d, %d) = %v, want %v", tc.zeros, tc.ones, got, tc.want)
		}
	}
}

func TestTree_SelectsUsefulFeatureAndPredictsThreshold(t *testing.T) {
	samples := []Sample{
		{[]float64{7, 4}, 1},
		{[]float64{7, 1}, 0},
		{[]float64{7, 3}, 1},
		{[]float64{7, 2}, 0},
	}
	tree, err := TrainTree(samples, 1)
	if err != nil {
		t.Fatal(err)
	}
	if tree.root.feature != 1 || tree.root.threshold != 2.5 {
		t.Fatalf("expected split on feature 1 at 2.5, got %+v", tree.root)
	}
	for _, tc := range []Sample{
		{[]float64{7, 0}, 0},
		{[]float64{7, 2.5}, 0},
		{[]float64{7, 2.6}, 1},
		{[]float64{7, 5}, 1},
	} {
		got, err := tree.Predict(tc.Features)
		if err != nil || got != tc.Label {
			t.Errorf("Predict(%v) = %d, %v; want %d", tc.Features, got, err, tc.Label)
		}
	}
	if samples[0].Features[1] != 4 || samples[1].Features[1] != 1 {
		t.Fatal("training changed the input order")
	}
	again, err := TrainTree(samples, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tree, again) {
		t.Fatal("training the same samples produced a different tree")
	}
}

func TestTree_WeightsGiniByChildSize(t *testing.T) {
	samples := []Sample{
		{[]float64{1}, 0}, {[]float64{2}, 0}, {[]float64{3}, 1},
		{[]float64{4}, 0}, {[]float64{5}, 1},
	}
	tree, err := TrainTree(samples, 1)
	if err != nil {
		t.Fatal(err)
	}
	// En 2.5: Gini ponderado = 4/15; en 4.5: 3/10.
	// Promediar sin ponderar elegiria incorrectamente el corte en 4.5.
	if tree.root.threshold != 2.5 {
		t.Fatalf("expected weighted Gini to select 2.5, got %v", tree.root.threshold)
	}
}

func TestTree_StoppingConditions(t *testing.T) {
	for _, tc := range []struct {
		name     string
		samples  []Sample
		maxDepth int
		want     int
	}{
		{"pure", []Sample{{[]float64{1}, 1}, {[]float64{2}, 1}}, 5, 1},
		{"one sample", []Sample{{[]float64{1}, 0}}, 5, 0},
		{"depth limit", []Sample{{[]float64{1}, 0}, {[]float64{2}, 1}, {[]float64{3}, 1}}, 0, 1},
		{"equal features", []Sample{{[]float64{2}, 1}, {[]float64{2}, 0}, {[]float64{2}, 1}}, 5, 1},
		{"majority zero", []Sample{{[]float64{2}, 0}, {[]float64{2}, 1}, {[]float64{2}, 0}}, 5, 0},
		{"tie", []Sample{{[]float64{2}, 1}, {[]float64{2}, 0}}, 5, 0},
		{"no improvement", []Sample{
			{[]float64{0, 0}, 0}, {[]float64{0, 1}, 1},
			{[]float64{1, 0}, 1}, {[]float64{1, 1}, 0},
		}, 5, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := TrainTree(tc.samples, tc.maxDepth)
			if err != nil {
				t.Fatal(err)
			}
			if tree.root.left != nil || tree.root.right != nil {
				t.Fatal("expected a leaf")
			}
			got, err := tree.Predict(tc.samples[0].Features)
			if err != nil || got != tc.want {
				t.Fatalf("prediction = %d, %v; want %d", got, err, tc.want)
			}
		})
	}
}

func TestTree_RecursiveSplitsRespectMaxDepth(t *testing.T) {
	samples := []Sample{
		{[]float64{1}, 0}, {[]float64{2}, 0},
		{[]float64{3}, 1}, {[]float64{4}, 1},
		{[]float64{5}, 0}, {[]float64{6}, 0},
	}
	shallow, err := TrainTree(samples, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, child := range []*node{shallow.root.left, shallow.root.right} {
		if child == nil || child.left != nil || child.right != nil {
			t.Fatal("expected both children to be leaves at depth 1")
		}
	}
	deep, err := TrainTree(samples, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range samples {
		got, err := deep.Predict(sample.Features)
		if err != nil || got != sample.Label {
			t.Errorf("Predict(%v) = %d, %v; want %d", sample.Features, got, err, sample.Label)
		}
	}
}

func TestTree_DuplicateValuesStayTogether(t *testing.T) {
	samples := []Sample{
		{[]float64{1}, 0}, {[]float64{1}, 1},
		{[]float64{2}, 1}, {[]float64{2}, 1},
	}
	tree, err := TrainTree(samples, 3)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []Sample{{[]float64{1}, 0}, {[]float64{2}, 1}} {
		got, err := tree.Predict(tc.Features)
		if err != nil || got != tc.Label {
			t.Errorf("Predict(%v) = %d, %v; want %d", tc.Features, got, err, tc.Label)
		}
	}
}

func TestTrainTree_InvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		samples  []Sample
		maxDepth int
	}{
		{"empty", nil, 1},
		{"no features", []Sample{{nil, 0}}, 1},
		{"invalid label", []Sample{{[]float64{1}, 2}}, 1},
		{"different lengths", []Sample{{[]float64{1}, 0}, {[]float64{1, 2}, 1}}, 1},
		{"NaN", []Sample{{[]float64{math.NaN()}, 0}}, 1},
		{"infinity", []Sample{{[]float64{math.Inf(1)}, 0}}, 1},
		{"negative depth", []Sample{{[]float64{1}, 0}}, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := TrainTree(tc.samples, tc.maxDepth); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestPredict_InvalidInput(t *testing.T) {
	tree, err := TrainTree([]Sample{{[]float64{1}, 0}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, features := range [][]float64{nil, {1, 2}, {math.NaN()}, {math.Inf(-1)}} {
		if _, err := tree.Predict(features); err == nil {
			t.Errorf("expected an error for features %v", features)
		}
	}
	for _, untrained := range []*DecisionTree{nil, {}} {
		if _, err := untrained.Predict([]float64{1}); err == nil {
			t.Fatal("expected an error for an untrained tree")
		}
	}
}
