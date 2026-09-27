// Package ml implementa modelos de clasificacion con la libreria estandar.
package ml

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// Sample contiene las features numericas y la clase: 0 (normal) o 1 (fraude).
// Todas las muestras deben usar las mismas features en el mismo orden.
type Sample struct {
	Features []float64
	Label    int
}

// El orden de estas columnas es el orden de entrada al modelo.
var featureColumns = [...]string{
	"typeCode", "amount", "oldbalanceOrg", "newbalanceOrig",
	"oldbalanceDest", "newbalanceDest", "errorBalanceOrig",
	"errorBalanceDest", "isMerchantDest", "hour",
}

// LoadCleanCSV carga secuencialmente las diez features y el target isFraud.
// limit=0 lee todos los registros; un limite positivo toma los primeros N.
// No recalcula features ni utiliza las otras columnas del preprocessing.
func LoadCleanCSV(path string, limit int) ([]Sample, error) {
	if limit < 0 {
		return nil, fmt.Errorf("limit no puede ser negativo")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("leer header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for i, name := range header {
		name = strings.TrimSpace(name)
		if _, exists := columns[name]; exists {
			return nil, fmt.Errorf("columna duplicada en header: %s", name)
		}
		columns[name] = i
	}
	var indices [len(featureColumns)]int
	for i, name := range featureColumns {
		index, exists := columns[name]
		if !exists {
			return nil, fmt.Errorf("falta columna requerida: %s", name)
		}
		indices[i] = index
	}
	labelIndex, exists := columns["isFraud"]
	if !exists {
		return nil, fmt.Errorf("falta columna requerida: isFraud")
	}

	var samples []Sample
	for limit == 0 || len(samples) < limit {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("leer CSV: %w", err)
		}
		line, _ := reader.FieldPos(0)
		features := make([]float64, len(featureColumns))
		for i, column := range indices {
			value, err := strconv.ParseFloat(strings.TrimSpace(row[column]), 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("fila %d, columna %s: numero invalido %q", line, featureColumns[i], row[column])
			}
			features[i] = value
		}
		label, err := strconv.Atoi(strings.TrimSpace(row[labelIndex]))
		if err != nil || (label != 0 && label != 1) {
			return nil, fmt.Errorf("fila %d, columna isFraud: se esperaba 0 o 1, se recibio %q", line, row[labelIndex])
		}
		samples = append(samples, Sample{Features: features, Label: label})
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("el CSV no contiene muestras")
	}
	return samples, nil
}

// StratifiedSplit divide aproximadamente 80% training / 20% test por clase.
// El 80/20 es una decision tecnica del proyecto, no un requisito del profesor.
// Se redondea hacia abajo el 80% por clase. Cada clase presente debe tener
// al menos dos muestras para aparecer en ambos conjuntos. No se balancean clases.
// Las filas se asignan una sola vez y las features se comparten solo para lectura.
func StratifiedSplit(samples []Sample, seed int64) (train, test []Sample, err error) {
	if _, err := validateTraining(samples, 0); err != nil {
		return nil, nil, err
	}
	var classes [2][]Sample
	for _, sample := range samples {
		classes[sample.Label] = append(classes[sample.Label], sample)
	}
	rng := rand.New(rand.NewSource(seed))
	for label, group := range classes {
		if len(group) == 0 {
			continue
		}
		if len(group) == 1 {
			return nil, nil, fmt.Errorf("clase %d: se necesitan al menos dos muestras para dividir training/test; use mas registros", label)
		}
		rng.Shuffle(len(group), func(i, j int) { group[i], group[j] = group[j], group[i] })
		cut := len(group) * 4 / 5
		train = append(train, group[:cut]...)
		test = append(test, group[cut:]...)
	}
	// Mezclar tambien el orden final evita dejar las filas agrupadas por label.
	rng.Shuffle(len(train), func(i, j int) { train[i], train[j] = train[j], train[i] })
	rng.Shuffle(len(test), func(i, j int) { test[i], test[j] = test[j], test[i] })
	return train, test, nil
}
