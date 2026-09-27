package benchmark

import (
	"syscall"
	"time"
)

// processCPUTime devuelve el tiempo de CPU (usuario + kernel) consumido por
// todo el proceso, sumando todos los hilos del sistema que usa Go.
func processCPUTime() (time.Duration, error) {
	var creation, exit, kernel, user syscall.Filetime
	handle, err := syscall.GetCurrentProcess()
	if err != nil {
		return 0, err
	}
	if err := syscall.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		return 0, err
	}
	return filetimeDuration(kernel) + filetimeDuration(user), nil
}

// Un Filetime cuenta intervalos de 100 ns.
func filetimeDuration(ft syscall.Filetime) time.Duration {
	return time.Duration(uint64(ft.HighDateTime)<<32|uint64(ft.LowDateTime)) * 100
}
