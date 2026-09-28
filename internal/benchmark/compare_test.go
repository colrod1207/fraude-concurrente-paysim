package benchmark

import (
	"testing"
	"time"
)

func TestCompare_SpeedupAgainstSequential(t *testing.T) {
	cfg := Config{Runs: 3, Trim: 0, Workers: []int{1, 4}}
	var seqCalls int
	calls := map[int]int{}
	results, err := Compare("demo", cfg,
		func() error { seqCalls++; time.Sleep(40 * time.Millisecond); return nil },
		func(w int) error { calls[w]++; time.Sleep(time.Duration(40/w) * time.Millisecond); return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	// 3 corridas cronometradas + 1 de perfil por configuracion.
	if seqCalls != 4 || calls[1] != 4 || calls[4] != 4 {
		t.Fatalf("llamadas inesperadas: secuencial=%d concurrente=%v", seqCalls, calls)
	}
	if len(results) != 3 || results[0].Mode != "secuencial" || results[0].Speedup != 1 {
		t.Fatalf("la primera fila debe ser la secuencial con speedup 1: %+v", results)
	}
	if r := results[2]; r.Workers != 4 || r.Speedup < 2 || r.Efficiency != r.Speedup/4 {
		t.Fatalf("con 4 workers se esperaba speedup ~4: %+v", r)
	}
}

func TestCompare_InvalidWorkers(t *testing.T) {
	noop := func() error { return nil }
	noopW := func(int) error { return nil }
	for _, workers := range [][]int{nil, {0}, {2, -1}} {
		if _, err := Compare("demo", Config{Runs: 1, Workers: workers}, noop, noopW); err == nil {
			t.Fatalf("workers=%v: expected an error", workers)
		}
	}
}
