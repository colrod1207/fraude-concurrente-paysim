package preprocessing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSummary_WriteJSON(t *testing.T) {
	s := NewSummary()
	s.TotalRead = 13
	s.Valid = 7
	s.Discarded = 6
	s.DiscardedByReason[EmptyRow] = 1
	s.DiscardedByReason[WrongColumnCount] = 1
	s.Workers = 4
	s.ElapsedSeconds = 0.5

	path := filepath.Join(t.TempDir(), "resumen.json")
	if err := s.WriteJSON(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded["total_leidas"].(float64) != 13 {
		t.Errorf("expected total_leidas 13, got %v", decoded["total_leidas"])
	}
	if decoded["validas"].(float64) != 7 {
		t.Errorf("expected validas 7, got %v", decoded["validas"])
	}
	if decoded["descartadas"].(float64) != 6 {
		t.Errorf("expected descartadas 6, got %v", decoded["descartadas"])
	}
	reasons, ok := decoded["descartadas_por_motivo"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected descartadas_por_motivo to be an object, got %T", decoded["descartadas_por_motivo"])
	}
	if reasons[EmptyRow].(float64) != 1 {
		t.Errorf("expected %s count 1, got %v", EmptyRow, reasons[EmptyRow])
	}
	if decoded["workers_usados"].(float64) != 4 {
		t.Errorf("expected workers_usados 4, got %v", decoded["workers_usados"])
	}
}
