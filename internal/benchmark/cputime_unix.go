//go:build unix

package benchmark

import (
	"syscall"
	"time"
)

// processCPUTime devuelve el tiempo de CPU (usuario + sistema) consumido por
// todo el proceso, sumando todos los hilos del sistema que usa Go.
func processCPUTime() (time.Duration, error) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, err
	}
	return time.Duration(usage.Utime.Nano() + usage.Stime.Nano()), nil
}
