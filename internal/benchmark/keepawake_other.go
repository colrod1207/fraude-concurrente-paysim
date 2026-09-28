//go:build !windows

package benchmark

// KeepAwake no hace nada fuera de Windows; ahi se recomienda desactivar la
// suspension del equipo manualmente mientras corre el benchmark.
func KeepAwake() (release func()) { return func() {} }
