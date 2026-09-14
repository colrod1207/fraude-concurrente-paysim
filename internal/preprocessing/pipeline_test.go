package preprocessing

import (
	"os"
	"path/filepath"
	"testing"
)

const syntheticCSV = `step,type,amount,nameOrig,oldbalanceOrg,newbalanceOrig,nameDest,oldbalanceDest,newbalanceDest,isFraud,isFlaggedFraud
1,PAYMENT,9839.64,C1231006815,170136.0,160296.36,M1979787155,0.0,0.0,0,0
1,TRANSFER,181.0,C1305486145,181.0,0.0,C553264065,0.0,0.0,1,0
1,CASH_OUT,181.0,C840083671,181.0,0.0,C38997010,21182.0,0.0,1,0
1,CASH_IN,5000.0,C900000001,1000.0,6000.0,M900000002,0.0,0.0,0,0
2,DEBIT,250.5,C900000003,5000.0,4749.5,C900000004,300.0,550.5,0,0
5,TRANSFER,3000000.0,C900000005,3000000.0,0.0,C900000006,0.0,3000000.0,1,1
25,CASH_OUT,50000.0,C900000007,50000.0,0.0,M900000008,10000.0,60000.0,0,0
,,,,,,,,,,
30,BOGUS_TYPE,100.0,C900000009,100.0,0.0,C900000010,0.0,100.0,0,0
40,PAYMENT,-50.0,C900000011,100.0,150.0,M900000012,0.0,0.0,0,0
50,PAYMENT,abc,C900000013,100.0,50.0,M900000014,0.0,0.0,0,0
60,PAYMENT,100.0,C900000015
70,CASH_IN,200.0,C900000016,-10.0,190.0,M900000017,0.0,0.0,0,0
`

func runPipelineOnSyntheticCSV(t *testing.T, numWorkers int) *Summary {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "clean_input.csv")
	output := filepath.Join(dir, "clean_output.csv")
	summaryPath := filepath.Join(dir, "resumen.json")

	if err := os.WriteFile(input, []byte(syntheticCSV), 0o644); err != nil {
		t.Fatalf("failed to write synthetic csv: %v", err)
	}

	summary, err := Run(input, output, summaryPath, numWorkers)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	return summary
}

func TestRun_SyntheticDataset_MatchesExpectedCounts(t *testing.T) {
	for _, workers := range []int{1, 2, 4} {
		summary := runPipelineOnSyntheticCSV(t, workers)

		if summary.TotalRead != 13 {
			t.Errorf("workers=%d: expected TotalRead 13, got %d", workers, summary.TotalRead)
		}
		if summary.Valid != 7 {
			t.Errorf("workers=%d: expected Valid 7, got %d", workers, summary.Valid)
		}
		if summary.Discarded != 6 {
			t.Errorf("workers=%d: expected Discarded 6, got %d", workers, summary.Discarded)
		}

		expectedReasons := map[string]int{
			EmptyRow:          1,
			WrongColumnCount:  1,
			ParseError:        1,
			InvalidType:       1,
			NegativeAmount:    1,
			ImpossibleBalance: 1,
		}
		for reason, want := range expectedReasons {
			if got := summary.DiscardedByReason[reason]; got != want {
				t.Errorf("workers=%d: expected %s count %d, got %d", workers, reason, want, got)
			}
		}
	}
}
