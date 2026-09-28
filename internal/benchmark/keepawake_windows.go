package benchmark

import (
	"runtime"
	"syscall"
)

// Banderas de SetThreadExecutionState (winbase.h).
const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

var setThreadExecutionState = syscall.NewLazyDLL("kernel32.dll").NewProc("SetThreadExecutionState")

// KeepAwake pide a Windows que no suspenda el equipo ni apague la pantalla
// hasta que se llame a la funcion devuelta. Sin esto, una laptop sin uso
// entra en "modo de espera moderno" a mitad del benchmark y baja la
// frecuencia de la CPU, lo que multiplica los tiempos medidos.
//
// La peticion pertenece al hilo del sistema que la hace, asi que una
// goroutine fija a su hilo la mantiene viva hasta que se libera.
func KeepAwake() (release func()) {
	if setThreadExecutionState.Find() != nil {
		return func() {}
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		setThreadExecutionState.Call(esContinuous | esSystemRequired | esDisplayRequired)
		<-stop
		setThreadExecutionState.Call(esContinuous)
		close(done)
	}()
	return func() {
		close(stop)
		<-done
	}
}
