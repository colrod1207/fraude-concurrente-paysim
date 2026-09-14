package preprocessing

import (
	"encoding/json"
	"os"
)

// Summary es la bitácora de limpieza (resumen_limpieza.json).
type Summary struct {
	TotalRead         int            `json:"total_leidas"`
	Valid             int            `json:"validas"`
	Discarded         int            `json:"descartadas"`
	DiscardedByReason map[string]int `json:"descartadas_por_motivo"`
	Workers           int            `json:"workers_usados"`
	ElapsedSeconds    float64        `json:"tiempo_segundos"`
}

// NewSummary crea un Summary listo para acumular resultados.
func NewSummary() *Summary {
	return &Summary{DiscardedByReason: make(map[string]int)}
}

// WriteJSON escribe el resumen en formato JSON indentado.
func (s *Summary) WriteJSON(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
