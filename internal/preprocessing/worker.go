package preprocessing

import (
	"strconv"
	"strings"
)

func isEmptyRow(fields []string) bool {
	if len(fields) == 0 {
		return true
	}
	for _, f := range fields {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}

// ProcessRow valida y transforma una fila cruda del CSV. Es una función
// pura (sin estado compartido) para que cada worker pueda invocarla en su
// propia goroutine sin necesidad de locks.
func ProcessRow(fields []string) Result {
	if isEmptyRow(fields) {
		return Result{Discard: EmptyRow}
	}
	if len(fields) != ExpectedColumns {
		return Result{Discard: WrongColumnCount}
	}

	step, errStep := strconv.Atoi(strings.TrimSpace(fields[0]))
	txType := strings.TrimSpace(fields[1])
	amount, errAmount := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
	nameOrig := strings.TrimSpace(fields[3])
	oldBalanceOrg, errOldOrg := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64)
	newBalanceOrig, errNewOrig := strconv.ParseFloat(strings.TrimSpace(fields[5]), 64)
	nameDest := strings.TrimSpace(fields[6])
	oldBalanceDest, errOldDest := strconv.ParseFloat(strings.TrimSpace(fields[7]), 64)
	newBalanceDest, errNewDest := strconv.ParseFloat(strings.TrimSpace(fields[8]), 64)
	isFraud, errIsFraud := strconv.Atoi(strings.TrimSpace(fields[9]))
	isFlaggedFraud, errIsFlagged := strconv.Atoi(strings.TrimSpace(fields[10]))

	if errStep != nil || errAmount != nil || errOldOrg != nil || errNewOrig != nil ||
		errOldDest != nil || errNewDest != nil || errIsFraud != nil || errIsFlagged != nil {
		return Result{Discard: ParseError}
	}

	typeCode, ok := ValidTypes[txType]
	if !ok {
		return Result{Discard: InvalidType}
	}

	if amount < 0 {
		return Result{Discard: NegativeAmount}
	}
	if oldBalanceOrg < 0 || newBalanceOrig < 0 || oldBalanceDest < 0 || newBalanceDest < 0 {
		return Result{Discard: ImpossibleBalance}
	}

	isMerchantDest := 0
	if strings.HasPrefix(nameDest, "M") {
		isMerchantDest = 1
	}

	record := &CleanRecord{
		Step:             step,
		Type:             txType,
		TypeCode:         typeCode,
		Amount:           amount,
		NameOrig:         nameOrig,
		OldBalanceOrg:    oldBalanceOrg,
		NewBalanceOrig:   newBalanceOrig,
		NameDest:         nameDest,
		OldBalanceDest:   oldBalanceDest,
		NewBalanceDest:   newBalanceDest,
		IsFraud:          isFraud,
		IsFlaggedFraud:   isFlaggedFraud,
		ErrorBalanceOrig: oldBalanceOrg - amount - newBalanceOrig,
		ErrorBalanceDest: oldBalanceDest + amount - newBalanceDest,
		IsMerchantDest:   isMerchantDest,
		Hour:             step % 24,
	}
	return Result{Record: record}
}
