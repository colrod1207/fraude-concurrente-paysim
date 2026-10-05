# fraude-concurrente — Detección de fraude concurrente con PaySim (PC1 y PC2)

Proyecto en Go (solo librería estándar, sin dependencias externas) del curso
CC65 (Programación Concurrente y Distribuida, UPC): limpieza concurrente del
dataset PaySim con el patrón productor/consumidor/reductor (goroutines,
channels y `sync.WaitGroup`), un Random Forest secuencial y concurrente
(worker pool), el modelo Promela de la sincronización y el cálculo de
Speedup con media recortada.

## Ejecución en un comando

Solo se necesita [Go](https://go.dev/dl/) 1.21 o superior (si la versión
instalada es más antigua que la del `go.mod`, Go descarga sola la correcta
al ejecutar el comando). No hace falta cuenta de Kaggle, token, Python ni
descargar nada a mano:

```bash
git clone https://github.com/colrod1207/fraude-concurrente-paysim.git
cd fraude-concurrente-paysim
go run ./cmd/benchmark
```

Ese comando hace todo el Entregable 2 de principio a fin:

1. Si no está el dataset, lo **descarga de Kaggle** automáticamente (dataset
   público, ~470 MB) en `data/raw/`.
2. Si no está el CSV limpio, **ejecuta la limpieza concurrente** y lo deja en
   `data/processed/`.
3. **Cronometra** la limpieza y el Random Forest en versión secuencial y
   concurrente (1, 2, 4 y 8 workers), 10 corridas cada una, y calcula la
   **media recortada**, el **Speedup** y el uso de CPU y memoria.
4. **Evalúa el Random Forest** (matriz de confusión, accuracy, precision,
   recall y F1).
5. Guarda todo en `results/` (`benchmark.md`, `benchmark.csv`,
   `metricas_random_forest.md`).

La corrida completa tarda unos 20 minutos. Para una prueba rápida (~5 min):
`go run ./cmd/benchmark --runs 2 --trim 0 --limit 50000 --trees 8`.

La descarga y la limpieza se hacen una sola vez; las siguientes ejecuciones
reutilizan los archivos. Los CSV están en `.gitignore`: nunca se suben a
GitHub. Cualquier otro comando (`go run .`, `go run ./cmd/train`) también
descarga el dataset si falta; `go run ./cmd/download` solo lo descarga.

> Si el repo está dentro de OneDrive/Dropbox, conviene descargar el dataset
> fuera para que no se sincronice: `go run ./cmd/download --dir C:/datasets/paysim`
> y pasar esa ruta con `--input` / `--raw`.

## Ejecutar la limpieza

```bash
go run . --workers 8
```

Por defecto lee `data/raw/PS_20174392719_1491204439457_log.csv` (y lo
descarga si falta). Las rutas se cambian con `--input`, `--output` y
`--summary`.

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
| Modelo Promela (limpieza) | `formal/sync_model.pml`, `formal/evidencia_spin.txt` |
| Modelo Promela (Random Forest) | `formal/forest_sync_model.pml`, `formal/evidencia_forest_spin.txt` |
| Análisis de código con IA (prompt y GAPs) | `docs/ai_code_review.md` |
| Integración continua | `.github/workflows/ci.yml` |

## Benchmark: Speedup y media recortada (PC2)

`cmd/benchmark` cronometra la versión secuencial y la concurrente de la
limpieza (`RunSequential` vs `Run`) y del Random Forest (`TrainForest` vs
`TrainForestConcurrent`) con 1, 2, 4 y 8 workers:

```bash
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

## Resultados (dataset real)

**Equipo:** laptop Intel Core i5-13420H (8 núcleos híbridos: 4 de rendimiento
+ 4 de eficiencia, 12 hilos), Windows 11, Go 1.27, conectada a la corriente,
modo de energía "Máximo rendimiento", proceso con prioridad alta y sin
suspensión del equipo durante la medición. **Método:** 10 corridas por
configuración, media recortada al 10 % por extremo (se descartan la corrida
más rápida y la más lenta). Tiempos de cada corrida en
[`results/benchmark.csv`](results/benchmark.csv).

### Limpieza (6 362 620 filas, 470 MB)

| Versión | Workers | Media recortada (s) | Speedup | Eficiencia | Núcleos usados |
|---|---:|---:|---:|---:|---:|
| Secuencial | 1 | 14.49 | 1.00 | 1.00 | 1.97 |
| Concurrente | 1 | 13.42 | 1.08 | 1.08 | 2.64 |
| Concurrente | 2 | 14.02 | 1.03 | 0.52 | 2.76 |
| Concurrente | 4 | 14.51 | 1.00 | 0.25 | 3.01 |
| Concurrente | 8 | 14.84 | 0.98 | 0.12 | 3.07 |

Resultado de la limpieza: 6 362 620 filas válidas, 0 descartadas
([`data/processed/resumen_limpieza.json`](data/processed/resumen_limpieza.json)).

### Random Forest (160 000 muestras de entrenamiento, 32 árboles, profundidad 8)

| Versión | Workers | Media recortada (s) | Speedup | Eficiencia | Núcleos usados |
|---|---:|---:|---:|---:|---:|
| Secuencial | 1 | 12.99 | 1.00 | 1.00 | 1.16 |
| Concurrente | 1 | 13.07 | 0.99 | 0.99 | 1.15 |
| Concurrente | 2 | 8.68 | 1.50 | 0.75 | 2.44 |
| Concurrente | 4 | 6.60 | 1.97 | 0.49 | 4.55 |
| Concurrente | 8 | 6.07 | 2.14 | 0.27 | 7.66 |

Calidad del modelo en test (40 001 muestras): accuracy 0.99998, precision
1.0000, recall 0.9667, F1 0.9831 (29 de 30 fraudes detectados, 0 falsas
alarmas). Detalle en
[`results/metricas_random_forest.md`](results/metricas_random_forest.md).

### Análisis

- **Random Forest: escala.** Entrenar un árbol es trabajo pesado de CPU e
  independiente de los demás, así que repartir árboles entre workers reduce
  el tiempo hasta ×2.14 con 8 workers. Con 1 worker el tiempo es igual al
  secuencial: el costo del pool (channel + WaitGroup) es despreciable. La
  eficiencia cae con más workers porque el entrenamiento recorre y ordena
  muchas muestras (más de 5 GB asignados en total), así que los workers
  compiten por la memoria y la caché, y porque 4 de los 8 núcleos son de
  eficiencia (~33 % más lentos).
- **Limpieza: no escala.** Validar una fila cuesta muy poco comparado con
  leerla del disco, pasarla por dos channels y escribirla. La lectura
  (productor) y la escritura del CSV (reductor) son una sola goroutine cada
  una, así que son la parte secuencial que, por la **ley de Amdahl**, limita
  el Speedup a ~1 sin importar cuántos workers se agreguen (los núcleos
  usados se quedan en ~3). La ganancia de ×1.08 con 1 worker viene de
  solapar lectura, validación y escritura en goroutines distintas (pipeline).
  Una mejora posible es enviar las filas en bloques para reducir el costo
  por mensaje.
- **Memoria:** el heap pico de la limpieza es de ~5 MB en todas las
  versiones porque el CSV se procesa en streaming (nunca se carga entero).
  En el Random Forest el heap pico crece con los workers (82 → 264 MB)
  porque cada worker mantiene su propia muestra bootstrap en memoria.

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
