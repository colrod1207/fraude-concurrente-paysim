// Descarga el CSV crudo de PaySim desde Kaggle (sin cuenta ni token) y lo
// deja en la ruta que usa por defecto la limpieza.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/colrod1207/fraude-concurrente-paysim/internal/dataset"
)

func main() {
	dir := flag.String("dir", "data/raw", "carpeta donde se guarda el CSV")
	url := flag.String("url", dataset.KaggleURL, "URL de descarga del .zip")
	flag.Parse()

	path, err := dataset.Download(*url, *dir, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Dataset listo -> %s\n", path)
}
