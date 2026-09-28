package ml

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const cleanTestCSV = `step,type,typeCode,amount,nameOrig,oldbalanceOrg,newbalanceOrig,nameDest,oldbalanceDest,newbalanceDest,isFraud,isFlaggedFraud,errorBalanceOrig,errorBalanceDest,isMerchantDest,hour
ignored,ignored,4,100,origin,200,90,destination,300,410,1,ignored,10,-10,0,1
ignored,ignored,3,25,origin,50,25,destination,0,0,0,ignored,0,25,1,2
ignored,ignored,1,50,origin,80,20,destination,10,70,1,ignored,10,-10,0,3
`

func writeCleanCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "clean.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadCleanCSV_FeatureOrderAndHeaderNames(t *testing.T) {
	want := []Sample{
		{[]float64{4, 100, 200, 90, 300, 410, 10, -10, 0, 1}, 1},
		{[]float64{3, 25, 50, 25, 0, 0, 0, 25, 1, 2}, 0},
		{[]float64{1, 50, 80, 20, 10, 70, 10, -10, 0, 3}, 1},
	}
	rows, err := csv.NewReader(strings.NewReader(cleanTestCSV)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	// Invertir todas las columnas comprueba que el lector usa sus nombres.
	for _, row := range rows {
		for i, j := 0, len(row)-1; i < j; i, j = i+1, j-1 {
			row[i], row[j] = row[j], row[i]
		}
	}
	var reversed strings.Builder
	if err := csv.NewWriter(&reversed).WriteAll(rows); err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{cleanTestCSV, reversed.String()} {
		got, err := LoadCleanCSV(writeCleanCSV(t, content), 0)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("loaded samples = %v, want %v", got, want)
		}
	}
}

func TestLoadCleanCSV_Limit(t *testing.T) {
	path := writeCleanCSV(t, cleanTestCSV)
	for _, tc := range []struct{ limit, want int }{{0, 3}, {1, 1}, {2, 2}, {10, 3}} {
		got, err := LoadCleanCSV(path, tc.limit)
		if err != nil || len(got) != tc.want {
			t.Fatalf("limit %d: got %d samples, %v; want %d", tc.limit, len(got), err, tc.want)
		}
		if got[0].Features[1] != 100 {
			t.Fatal("limit must preserve the first rows without random sampling")
		}
	}
	if _, err := LoadCleanCSV(path, -1); err == nil {
		t.Fatal("expected an error for a negative limit")
	}
}

func TestLoadCleanCSV_InvalidInput(t *testing.T) {
	header := strings.SplitN(cleanTestCSV, "\n", 2)[0] + "\n"
	for _, tc := range []struct {
		name, content, message string
	}{
		{"missing feature", strings.Replace(cleanTestCSV, "amount", "other", 1), "amount"},
		{"missing target", strings.Replace(cleanTestCSV, "isFraud", "other", 1), "isFraud"},
		{"duplicate header", strings.Replace(cleanTestCSV, "step", "amount", 1), "duplicada"},
		{"invalid number", strings.Replace(cleanTestCSV, ",100,", ",abc,", 1), "fila 2, columna amount"},
		{"empty number", strings.Replace(cleanTestCSV, ",100,", ",,", 1), "amount"},
		{"NaN", strings.Replace(cleanTestCSV, ",100,", ",NaN,", 1), "amount"},
		{"infinity", strings.Replace(cleanTestCSV, ",100,", ",+Inf,", 1), "amount"},
		{"overflow", strings.Replace(cleanTestCSV, ",100,", ",1e999,", 1), "amount"},
		{"invalid label", strings.Replace(cleanTestCSV, ",1,ignored,", ",2,ignored,", 1), "isFraud"},
		{"non numeric label", strings.Replace(cleanTestCSV, ",1,ignored,", ",abc,ignored,", 1), "isFraud"},
		{"wrong column count", header + "1,2\n", "leer CSV"},
		{"empty file", "", "header"},
		{"no samples", header, "no contiene muestras"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadCleanCSV(writeCleanCSV(t, tc.content), 0)
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("error = %v, want message containing %q", err, tc.message)
			}
		})
	}
	if _, err := LoadCleanCSV(filepath.Join(t.TempDir(), "missing.csv"), 0); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestStratifiedSplit_ProportionsDeterminismAndNoOverlap(t *testing.T) {
	samples := make([]Sample, 20)
	for i := range samples {
		label := 0
		if i >= 15 {
			label = 1
		}
		// El valor identifica la fila para comprobar que se asigna una sola vez.
		samples[i] = Sample{Features: []float64{float64(i)}, Label: label}
	}
	train, test, err := StratifiedSplit(samples, 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(train) != 16 || len(test) != 4 {
		t.Fatalf("got train/test sizes %d/%d, want 16/4", len(train), len(test))
	}
	seen := make(map[float64]bool)
	for i, group := range [][]Sample{train, test} {
		var counts [2]int
		for _, sample := range group {
			id := sample.Features[0]
			if seen[id] {
				t.Fatalf("row %v appears more than once in train/test", id)
			}
			seen[id] = true
			counts[sample.Label]++
		}
		want := [][2]int{{12, 4}, {3, 1}}[i]
		if counts != want {
			t.Fatalf("group %d: class counts = %v, want %v", i, counts, want)
		}
	}
	for i, sample := range samples {
		if sample.Features[0] != float64(i) || !seen[float64(i)] {
			t.Fatalf("input row %d was changed or omitted", i)
		}
	}
	trainAgain, testAgain, err := StratifiedSplit(samples, 42)
	if err != nil || !reflect.DeepEqual(train, trainAgain) || !reflect.DeepEqual(test, testAgain) {
		t.Fatalf("same seed did not reproduce the split: %v", err)
	}
}

func TestStratifiedSplit_SmallClasses(t *testing.T) {
	samples := []Sample{
		{[]float64{0}, 0}, {[]float64{1}, 0}, {[]float64{2}, 0},
		{[]float64{3}, 1}, {[]float64{4}, 1},
	}
	train, test, err := StratifiedSplit(samples, 42)
	if err != nil || len(train) != 3 || len(test) != 2 {
		t.Fatalf("expected floor(0.8 * class size): train=%d, test=%d, err=%v", len(train), len(test), err)
	}
	for _, group := range [][]Sample{train, test} {
		ones := 0
		for _, sample := range group {
			ones += sample.Label
		}
		if ones != 1 {
			t.Fatal("both groups must contain one row of the minority class")
		}
	}
	// Una clase ausente no se inventa ni se balancea.
	if _, _, err := StratifiedSplit(samples[:3], 42); err != nil {
		t.Fatal(err)
	}
}

func TestStratifiedSplit_InvalidInput(t *testing.T) {
	for _, samples := range [][]Sample{
		nil,
		{{[]float64{0}, 0}},
		{{[]float64{0}, 0}, {[]float64{1}, 0}, {[]float64{2}, 1}},
		{{[]float64{0}, 2}},
		{{nil, 0}},
	} {
		if _, _, err := StratifiedSplit(samples, 42); err == nil {
			t.Fatalf("expected an error for %v", samples)
		}
	}
}
