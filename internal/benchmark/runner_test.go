package benchmark

import (
	"errors"
	"testing"
	"time"
)

func TestTime_RunsEachTime(t *testing.T) {
	calls := 0
	runs, err := Time(3, func() error { calls++; return nil })
	if err != nil || len(runs) != 3 || calls != 3 {
		t.Fatalf("Time = %v, %v con %d llamadas; want 3 corridas", runs, err, calls)
	}
	if _, err := Time(0, func() error { return nil }); err == nil {
		t.Fatal("expected an error for zero runs")
	}
	boom := errors.New("boom")
	if _, err := Time(2, func() error { return boom }); !errors.Is(err, boom) {
		t.Fatalf("Time debe propagar el error, got %v", err)
	}
}

func TestProfileRun_MeasuresCPUAndMemory(t *testing.T) {
	var sink [][]byte
	profile, err := ProfileRun(func() error {
		deadline := time.Now().Add(50 * time.Millisecond)
		for time.Now().Before(deadline) {
			sink = append(sink, make([]byte, 64<<10))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sink) == 0 {
		t.Fatal("la funcion debio ejecutarse")
	}
	if profile.Wall < 50*time.Millisecond {
		t.Fatalf("Wall = %v, want >= 50ms", profile.Wall)
	}
	if profile.AllocMB <= 0 || profile.PeakHeapMB <= 0 {
		t.Fatalf("se esperaba memoria positiva: %+v", profile)
	}
	if profile.CPU <= 0 || profile.CPUPercent <= 0 {
		t.Fatalf("se esperaba uso de CPU positivo: %+v", profile)
	}
}
