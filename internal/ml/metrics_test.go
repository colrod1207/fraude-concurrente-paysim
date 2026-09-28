package ml

import (
	"math"
	"testing"
)

func TestCalculateMetrics_ConfusionMatrixAndRates(t *testing.T) {
	actual := []int{1, 1, 0, 0, 0, 0, 1, 1}
	predicted := []int{1, 1, 0, 0, 0, 1, 0, 0}
	got, err := CalculateMetrics(actual, predicted)
	if err != nil {
		t.Fatal(err)
	}
	if got.TP != 2 || got.TN != 3 || got.FP != 1 || got.FN != 2 {
		t.Fatalf("incorrect confusion matrix: %+v", got)
	}
	for _, tc := range []struct {
		name      string
		got, want float64
	}{
		{"accuracy", got.Accuracy, 5.0 / 8},
		{"precision", got.Precision, 2.0 / 3},
		{"recall", got.Recall, 2.0 / 4},
		{"F1", got.F1, 4.0 / 7},
	} {
		if math.Abs(tc.got-tc.want) > 1e-12 {
			t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}

func TestCalculateMetrics_ZeroDenominators(t *testing.T) {
	for _, tc := range []struct {
		name              string
		actual, predicted []int
		want              Metrics
	}{
		{"empty", nil, nil, Metrics{}},
		{"only negatives", []int{0, 0}, []int{0, 0}, Metrics{TN: 2, Accuracy: 1}},
		{"no predicted positives", []int{1, 0}, []int{0, 0}, Metrics{TN: 1, FN: 1, Accuracy: 0.5}},
		{"no actual positives", []int{0, 0}, []int{1, 0}, Metrics{TN: 1, FP: 1, Accuracy: 0.5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CalculateMetrics(tc.actual, tc.predicted)
			if err != nil || got != tc.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
}

func TestCalculateMetrics_InvalidInput(t *testing.T) {
	for _, tc := range []struct{ actual, predicted []int }{
		{[]int{0}, nil}, {[]int{2}, []int{0}}, {[]int{0}, []int{-1}},
	} {
		if _, err := CalculateMetrics(tc.actual, tc.predicted); err == nil {
			t.Fatal("expected an error")
		}
	}
}
