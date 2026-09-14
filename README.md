# fraude-concurrente — Limpieza concurrente del dataset PaySim (PC1)

Pipeline en Go (solo librería estándar, sin dependencias externas) que
carga, limpia y valida el dataset PaySim usando goroutines, channels y
`sync.WaitGroup`, siguiendo el patrón productor/consumidor/reductor
descrito en el informe PC1 del curso CC65 (Programación Concurrente y
Distribuida, UPC).

## Descargar el dataset

Vía API de Kaggle (recomendado — cada integrante necesita su propia cuenta
y token, no se comparte `kaggle.json` entre personas):

```bash
pip install kaggle
kaggle datasets download -d ealaxi/paysim1 -p data/raw --unzip
```

Requiere tener el token de tu cuenta en `~/.kaggle/kaggle.json` (se genera
en https://www.kaggle.com/settings → API → "Create New Token").

## Ejecutar la limpieza

```bash
go run . \
  --input data/raw/PS_20174392719_1491204439457_log.csv \
  --output data/processed/paysim_clean.csv \
  --summary data/processed/resumen_limpieza.json \
  --workers 8
```

Esto genera:

- `data/processed/paysim_clean.csv`: dataset limpio con las variables
  derivadas (`errorBalanceOrig`, `errorBalanceDest`, `isMerchantDest`,
  `hour`, `typeCode`).
- `data/processed/resumen_limpieza.json`: bitácora con el conteo de filas
  leídas, válidas, descartadas y el motivo de cada descarte. **Este sí se
  sube** como evidencia de la limpieza (Entregable 1, rúbrica: "Resultado
  dataset limpio").

## Arquitectura concurrente

```
goroutine productora                  pool de N goroutines worker
─────────────────────                 ──────────────────────────
StreamCSV() lee el CSV       -----> ProcessRow() valida,
línea a línea (sin cargarlo         transforma y calcula
entero a memoria) y lo envía        las variables derivadas
por un channel de filas                     │
        │                                   │
        └──────────── channel Result ───────┘
        │
        ▼
goroutine reductora (la que llama a Run):
escribe el CSV limpio y acumula el resumen
(único lugar que toca ese estado -> sin locks)
```

- Cada worker recibe una fila por el channel de entrada y devuelve un
  `Result` por el channel de salida, sin tocar ningún estado compartido —
  por eso no hay condición de carrera posible entre workers, por
  construcción (verificado también con `go test -race`).
- Solo la goroutine que llama a `Run` (el reductor) lee del channel de
  resultados, escribe el CSV de salida y acumula los contadores del
  resumen — al ser la única, tampoco hace falta ningún lock ahí.
- Un `sync.WaitGroup` señala cuándo todos los workers terminaron, para
  poder cerrar el channel de resultados y que el reductor deje de leer.
- El número de workers es un flag (`--workers`) pensado para el benchmark
  de Speedup del Entregable 2 (probar 1, 2, 4, 8...).
- `StreamCSV` lee el archivo línea a línea (nunca lo carga entero en
  memoria) y lo reparte por un channel con buffer, para no perder
  rendimiento serializando fila por fila.

## Estructura del proyecto

| Módulo | Archivos |
|---|---|
| Esquema + lector concurrente (productor) | `internal/preprocessing/schema.go`, `internal/preprocessing/reader.go` (+ tests) |
| Validación y transformación (worker) | `internal/preprocessing/worker.go` (+ test) |
| Orquestación, resumen y CLI | `internal/preprocessing/pipeline.go`, `internal/preprocessing/summary.go`, `internal/preprocessing/testdata.go`, `cmd/testdatagen/main.go`, `main.go` (+ tests) |

## Pruebas

El proyecto sigue TDD: cada archivo de `internal/preprocessing` tiene su
`_test.go` correspondiente (`go test ./...`), incluyendo un test de
integración que corre el pipeline completo sobre un CSV sintético con
casos borde deliberados (fila vacía, número de columnas incorrecto, error
de parseo, `type` inválido, monto negativo y saldo negativo) y verifica
que dé exactamente 13 filas leídas, 7 válidas y 6 descartadas (una por
motivo), con 1, 2 y 4 workers.

```bash
go test ./... -race
```

También se puede generar el mismo CSV sintético y correr el binario
manualmente:

```bash
go run ./cmd/testdatagen data/testdata/clean_input.csv
go run . --input data/testdata/clean_input.csv \
  --output data/testdata/clean_output.csv \
  --summary data/testdata/resumen.json --workers 4
```
