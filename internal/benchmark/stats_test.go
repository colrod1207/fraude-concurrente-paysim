package benchmark

import (
	"math"
	"testing"
	"time"
)

func TestTrimmedMean_DropsExtremes(t *testing.T) {
	// 10 corridas con 10% de recorte: se descartan la mas rapida (1s) y la
	// mas lenta (100s); el resto promedia 5s.
	runs := []time.Duration{
		5 * time.Second, 100 * time.Second, 4 * time.Second, 6 * time.Second,
		5 * time.Second, 1 * time.Second, 5 * time.Second, 4 * time.Second,
		6 * time.Second, 5 * time.Second,
	}
	got, err := TrimmedMean(runs, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	if got != 5*time.Second {
		t.Fatalf("TrimmedMean = %v, want 5s", got)
	}
}

func TestTrimmedMean_DoesNotModifyInput(t *testing.T) {
	runs := []time.Duration{3, 1, 2}
	if _, err := TrimmedMean(runs, 0.34); err != nil {
		t.Fatal(err)
	}
	if runs[0] != 3 || runs[1] != 1 || runs[2] != 2 {
		t.Fatalf("TrimmedMean no debe reordenar la entrada: %v", runs)
	}
}

func TestTrimmedMean_ZeroTrimIsPlainMean(t *testing.T) {
	got, err := TrimmedMean([]time.Duration{1, 2, 6}, 0)
	if err != nil || got != 3 {
		t.Fatalf("TrimmedMean = %v, %v; want 3", got, err)
	}
}

func TestTrimmedMean_InvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		runs []time.Duration
		trim float64
	}{
		{"empty", nil, 0.1},
		{"negative trim", []time.Duration{1, 2}, -0.1},
		{"trim 0.5", []time.Duration{1, 2}, 0.5},
		{"NaN trim", []time.Duration{1, 2}, math.NaN()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := TrimmedMean(tc.runs, tc.trim); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestSpeedupAndEfficiency(t *testing.T) {
	speedup := Speedup(8*time.Second, 2*time.Second)
	if speedup != 4 {
		t.Fatalf("Speedup = %v, want 4", speedup)
	}
	if eff := Efficiency(speedup, 8); eff != 0.5 {
		t.Fatalf("Efficiency = %v, want 0.5", eff)
	}
	if Speedup(time.Second, 0) != 0 || Efficiency(2, 0) != 0 {
		t.Fatal("denominadores cero deben devolver 0")
	}
}
