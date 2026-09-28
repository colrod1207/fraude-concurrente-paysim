package benchmark

import (
	"runtime"
	"testing"
)

func TestSetThreadExecutionState_IsHonored(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	want := uintptr(esContinuous | esSystemRequired | esDisplayRequired)
	if r, _, _ := setThreadExecutionState.Call(want); r == 0 {
		t.Fatal("Windows rechazo SetThreadExecutionState")
	}
	// La siguiente llamada devuelve el estado anterior del hilo: debe ser el pedido.
	if prev, _, _ := setThreadExecutionState.Call(esContinuous); prev != want {
		t.Fatalf("estado anterior = %#x, want %#x", prev, want)
	}
}
