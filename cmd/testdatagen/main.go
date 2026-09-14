// Genera un CSV sintético con el mismo esquema de PaySim y casos borde
// deliberados. Uso: go run ./cmd/testdatagen data/testdata/clean_input.csv
package main

import (
	"fmt"
	"os"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/preprocessing"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "uso: testdatagen <ruta-salida.csv>")
		os.Exit(1)
	}
	if err := preprocessing.WriteSyntheticCSV(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "error generando csv: %v\n", err)
		os.Exit(1)
	}
}
