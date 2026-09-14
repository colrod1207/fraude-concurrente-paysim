package preprocessing

import (
	"encoding/csv"
	"io"
	"os"
)

// StreamCSV es la goroutine productora: lee el CSV fila por fila (nunca lo
// carga entero en memoria) y la envía por el channel rows. Cierra el
// channel al terminar, sea por EOF o por error.
//
// FieldsPerRecord se fija en -1 para que filas con un número de columnas
// distinto al esperado lleguen tal cual al channel (y sea process_row quien
// decida descartarlas), en vez de que el csv.Reader las rechace por su
// cuenta.
func StreamCSV(path string, rows chan<- []string) error {
	defer close(rows)

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	first := true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rows <- nil
			continue
		}
		if first {
			first = false
			continue
		}
		rows <- record
	}
	return nil
}
