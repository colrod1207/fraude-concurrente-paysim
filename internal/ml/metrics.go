package ml

import "fmt"

// Metrics usa fraude (label=1) como clase positiva. Las tasas estan entre 0 y 1.
type Metrics struct {
	TP, TN, FP, FN              int
	Accuracy, Precision, Recall float64
	F1                          float64
}

// CalculateMetrics compara etiquetas reales y predichas en el mismo orden.
// Si un denominador es cero, la metrica correspondiente se devuelve como 0.
// Dos listas vacias producen conteos y tasas iguales a cero.
func CalculateMetrics(actual, predicted []int) (Metrics, error) {
	if len(actual) != len(predicted) {
		return Metrics{}, fmt.Errorf("las etiquetas reales y predichas deben tener la misma longitud")
	}
	var result Metrics
	for i, label := range actual {
		prediction := predicted[i]
		if (label != 0 && label != 1) || (prediction != 0 && prediction != 1) {
			return Metrics{}, fmt.Errorf("posicion %d: las etiquetas deben ser 0 o 1", i)
		}
		switch {
		case label == 1 && prediction == 1:
			result.TP++
		case label == 0 && prediction == 0:
			result.TN++
		case label == 0 && prediction == 1:
			result.FP++
		case label == 1 && prediction == 0:
			result.FN++
		}
	}
	result.Accuracy = ratio(result.TP+result.TN, len(actual))
	result.Precision = ratio(result.TP, result.TP+result.FP)
	result.Recall = ratio(result.TP, result.TP+result.FN)
	result.F1 = ratio(2*result.TP, 2*result.TP+result.FP+result.FN)
	return result, nil
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
