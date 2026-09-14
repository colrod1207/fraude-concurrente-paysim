package preprocessing

import "os"

const syntheticDatasetRows = `step,type,amount,nameOrig,oldbalanceOrg,newbalanceOrig,nameDest,oldbalanceDest,newbalanceDest,isFraud,isFlaggedFraud
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

// WriteSyntheticCSV genera un CSV con el mismo esquema de PaySim y casos
// borde deliberados (fila vacía, columnas incorrectas, error de parseo,
// type inválido, monto negativo, saldo negativo), usado para validar el
// pipeline sin depender del archivo real de 470 MB.
func WriteSyntheticCSV(path string) error {
	return os.WriteFile(path, []byte(syntheticDatasetRows), 0o644)
}
