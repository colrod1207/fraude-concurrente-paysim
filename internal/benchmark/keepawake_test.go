package benchmark

import "testing"

func TestKeepAwake_ReleaseReturns(t *testing.T) {
	release := KeepAwake()
	release() // no debe bloquearse ni fallar
}
