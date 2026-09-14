// Package preprocessing implementa la limpieza concurrente del dataset
// PaySim (PC1, CC65): una goroutine lectora (productor) entrega filas
// crudas del CSV por un channel, un pool de goroutines worker las valida y
// transforma en paralelo, y una única goroutine reductora escribe el CSV
// limpio y acumula el resumen -- al ser la única que toca ese estado, no
// hace falta ningún lock.
package preprocessing

// ExpectedColumns es el número de columnas de una fila válida del CSV
// crudo de PaySim.
const ExpectedColumns = 11

// Motivos de descarte de una fila.
const (
	EmptyRow          = "empty_row"
	WrongColumnCount  = "wrong_column_count"
	ParseError        = "parse_error"
	InvalidType       = "invalid_type"
	NegativeAmount    = "negative_amount"
	ImpossibleBalance = "impossible_balance"
)

// ValidTypes mapea cada tipo de transacción válido a su código ordinal.
var ValidTypes = map[string]int{
	"CASH_IN":  0,
	"CASH_OUT": 1,
	"DEBIT":    2,
	"PAYMENT":  3,
	"TRANSFER": 4,
}

// CleanRecord es una fila válida ya transformada, lista para escribirse en
// el CSV limpio.
type CleanRecord struct {
	Step             int
	Type             string
	TypeCode         int
	Amount           float64
	NameOrig         string
	OldBalanceOrg    float64
	NewBalanceOrig   float64
	NameDest         string
	OldBalanceDest   float64
	NewBalanceDest   float64
	IsFraud          int
	IsFlaggedFraud   int
	ErrorBalanceOrig float64
	ErrorBalanceDest float64
	IsMerchantDest   int
	Hour             int
}

// Result es lo que devuelve un worker por cada fila procesada: o un
// CleanRecord válido, o un motivo de descarte (nunca ambos).
type Result struct {
	Record  *CleanRecord
	Discard string
}

// Header devuelve las columnas del CSV limpio, en orden.
func Header() []string {
	return []string{
		"step", "type", "typeCode", "amount",
		"nameOrig", "oldbalanceOrg", "newbalanceOrig",
		"nameDest", "oldbalanceDest", "newbalanceDest",
		"isFraud", "isFlaggedFraud",
		"errorBalanceOrig", "errorBalanceDest",
		"isMerchantDest", "hour",
	}
}
