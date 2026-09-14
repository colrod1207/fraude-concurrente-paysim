package preprocessing

import "testing"

func TestProcessRow_EmptyRow(t *testing.T) {
	fields := []string{"", "", "", "", "", "", "", "", "", "", ""}
	result := ProcessRow(fields)
	if result.Discard != EmptyRow {
		t.Fatalf("expected discard %q, got %q (record=%v)", EmptyRow, result.Discard, result.Record)
	}
}

func TestProcessRow_NilRow(t *testing.T) {
	result := ProcessRow(nil)
	if result.Discard != EmptyRow {
		t.Fatalf("expected discard %q, got %q", EmptyRow, result.Discard)
	}
}

func TestProcessRow_WrongColumnCount(t *testing.T) {
	fields := []string{"60", "PAYMENT", "100.0", "C900000015"}
	result := ProcessRow(fields)
	if result.Discard != WrongColumnCount {
		t.Fatalf("expected discard %q, got %q", WrongColumnCount, result.Discard)
	}
}

func TestProcessRow_ParseError(t *testing.T) {
	fields := []string{"50", "PAYMENT", "abc", "C900000013", "100.0", "50.0", "M900000014", "0.0", "0.0", "0", "0"}
	result := ProcessRow(fields)
	if result.Discard != ParseError {
		t.Fatalf("expected discard %q, got %q", ParseError, result.Discard)
	}
}

func TestProcessRow_InvalidType(t *testing.T) {
	fields := []string{"30", "BOGUS_TYPE", "100.0", "C900000009", "100.0", "0.0", "C900000010", "0.0", "100.0", "0", "0"}
	result := ProcessRow(fields)
	if result.Discard != InvalidType {
		t.Fatalf("expected discard %q, got %q", InvalidType, result.Discard)
	}
}

func TestProcessRow_NegativeAmount(t *testing.T) {
	fields := []string{"40", "PAYMENT", "-50.0", "C900000011", "100.0", "150.0", "M900000012", "0.0", "0.0", "0", "0"}
	result := ProcessRow(fields)
	if result.Discard != NegativeAmount {
		t.Fatalf("expected discard %q, got %q", NegativeAmount, result.Discard)
	}
}

func TestProcessRow_ImpossibleBalance(t *testing.T) {
	fields := []string{"70", "CASH_IN", "200.0", "C900000016", "-10.0", "190.0", "M900000017", "0.0", "0.0", "0", "0"}
	result := ProcessRow(fields)
	if result.Discard != ImpossibleBalance {
		t.Fatalf("expected discard %q, got %q", ImpossibleBalance, result.Discard)
	}
}

func TestProcessRow_ValidRecordDerivedFields(t *testing.T) {
	fields := []string{"25", "CASH_OUT", "50000.0", "C900000007", "50000.0", "0.0", "M900000008", "10000.0", "60000.0", "0", "0"}
	result := ProcessRow(fields)
	if result.Discard != "" {
		t.Fatalf("expected no discard, got %q", result.Discard)
	}
	if result.Record == nil {
		t.Fatal("expected a record, got nil")
	}
	r := result.Record
	if r.TypeCode != 1 {
		t.Errorf("expected TypeCode 1 (CASH_OUT), got %d", r.TypeCode)
	}
	if r.ErrorBalanceOrig != 0.0 {
		t.Errorf("expected ErrorBalanceOrig 0.0 (50000 - 50000 - 0), got %f", r.ErrorBalanceOrig)
	}
	if r.ErrorBalanceDest != 0.0 {
		t.Errorf("expected ErrorBalanceDest 0.0 (10000 + 50000 - 60000), got %f", r.ErrorBalanceDest)
	}
	if r.IsMerchantDest != 1 {
		t.Errorf("expected IsMerchantDest 1 (nameDest starts with M), got %d", r.IsMerchantDest)
	}
	if r.Hour != 1 {
		t.Errorf("expected Hour 1 (25%%24), got %d", r.Hour)
	}
}
