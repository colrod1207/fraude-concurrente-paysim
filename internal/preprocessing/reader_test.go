package preprocessing

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp csv: %v", err)
	}
	return path
}

func TestStreamCSV_SkipsHeaderAndYieldsRows(t *testing.T) {
	content := "step,type,amount,nameOrig,oldbalanceOrg,newbalanceOrig,nameDest,oldbalanceDest,newbalanceDest,isFraud,isFlaggedFraud\n" +
		"1,PAYMENT,9839.64,C1231006815,170136.0,160296.36,M1979787155,0.0,0.0,0,0\n" +
		"1,TRANSFER,181.0,C1305486145,181.0,0.0,C553264065,0.0,0.0,1,0\n"
	path := writeTempCSV(t, content)

	rows := make(chan []string, 10)
	if err := StreamCSV(path, rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got [][]string
	for r := range rows {
		got = append(got, r)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 data rows (header skipped), got %d: %v", len(got), got)
	}
	if got[0][1] != "PAYMENT" {
		t.Errorf("expected first row type PAYMENT, got %q", got[0][1])
	}
	if got[1][1] != "TRANSFER" {
		t.Errorf("expected second row type TRANSFER, got %q", got[1][1])
	}
}

func TestStreamCSV_PreservesRowsWithWrongColumnCount(t *testing.T) {
	content := "header\n" +
		"60,PAYMENT,100.0,C900000015\n"
	path := writeTempCSV(t, content)

	rows := make(chan []string, 10)
	if err := StreamCSV(path, rows); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got [][]string
	for r := range rows {
		got = append(got, r)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if len(got[0]) != 4 {
		t.Fatalf("expected row to keep its 4 fields (not be rejected by the csv reader), got %d fields", len(got[0]))
	}
}
