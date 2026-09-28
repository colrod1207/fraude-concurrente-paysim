# fraude-concurrente — Limpieza concurrente del dataset PaySim (PC1)

Pipeline en Go (solo librería estándar, sin dependencias externas) que
carga, limpia y valida el dataset PaySim usando goroutines, channels y
`sync.WaitGroup`, siguiendo el patrón productor/consumidor/reductor
descrito en el informe PC1 del curso CC65 (Programación Concurrente y
Distribuida, UPC).

## Descargar el dataset

Un solo comando, sin cuenta de Kaggle, sin token y sin instalar nada más
que Go (el dataset es público):

```bash
go run ./cmd/download
```

Descarga el `.zip` desde la API de Kaggle, extrae
`data/raw/PS_20174392719_1491204439457_log.csv` (~470 MB) y borra el
`.zip`. Si el CSV ya existe no vuelve a descargarlo. Esa es la ruta que la
limpieza usa por defecto, así que después basta con `go run .`. El CSV está
en `.gitignore`: nunca se sube a GitHub.

> Si el repo está dentro de OneDrive/Dropbox, conviene descargarlo fuera
> para que no se sincronice: `go run ./cmd/download --dir C:/datasets/paysim`
> y pasar esa ruta con `--input`.

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
| Descarga del dataset | `internal/dataset/download.go`, `cmd/download/main.go` (+ test) |
| Esquema + lector concurrente (productor) | `internal/preprocessing/schema.go`, `internal/preprocessing/reader.go` (+ tests) |
| Validación y transformación (worker) | `internal/preprocessing/worker.go` (+ test) |
| Orquestación, resumen y CLI | `internal/preprocessing/pipeline.go`, `internal/preprocessing/summary.go`, `internal/preprocessing/testdata.go`, `cmd/testdatagen/main.go`, `main.go` (+ tests) |
| Limpieza secuencial (baseline) | `internal/preprocessing/pipeline_sequential.go` |
| Random Forest secuencial y concurrente | `internal/ml/`, `cmd/train/main.go` (+ tests) |
| Benchmark (Speedup, media recortada, CPU/memoria) | `internal/benchmark/`, `cmd/benchmark/main.go` (+ tests) |
| Modelo Promela | `formal/sync_model.pml`, `formal/evidencia_spin.txt` |

## Benchmark: Speedup y media recortada (PC2)

`cmd/benchmark` cronometra la versión secuencial y la concurrente de la
limpieza (`RunSequential` vs `Run`) y del Random Forest (`TrainForest` vs
`TrainForestConcurrent`) con 1, 2, 4 y 8 workers:

```bash
# 1) generar el CSV limpio (lo necesita el Random Forest)
go run . --input data/raw/PS_20174392719_1491204439457_log.csv --workers 8
# 2) cronometrar todo
go run ./cmd/benchmark --runs 10 --trim 0.1
```

- Cada configuración se ejecuta `--runs` veces; se ordenan los tiempos, se
  descarta el `--trim` (10%) más rápido y más lento, y se promedia el resto
  (**media recortada**).
- **Speedup** = T-Secuencial / T-Concurrente, y **eficiencia** =
  Speedup / workers.
- Aparte de las corridas cronometradas se hace una corrida extra para medir
  **CPU** (tiempo de CPU del proceso / tiempo real = núcleos usados) y
  **memoria** (heap pico y memoria total asignada), así esa medición no
  altera los tiempos.
- Del Random Forest solo se cronometra el entrenamiento (la carga del CSV y
  el split se hacen una vez). `--limit`, `--trees` y `--depth` controlan el
  tamaño del experimento.
- Los resultados quedan en `results/benchmark.md` (tabla para el informe) y
  `results/benchmark.csv` (incluye el tiempo de cada corrida).

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
