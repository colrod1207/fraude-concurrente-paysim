# fraude-concurrente — Limpieza concurrente del dataset PaySim (PC1)

Pipeline en Python (solo librería estándar, sin dependencias externas) que
carga, limpia y valida el dataset PaySim usando `multiprocessing`,
siguiendo el patrón productor/consumidor con un pool de procesos worker y
un proceso reductor, tal como se documenta en el informe PC1 del curso
CC65.

Se usa `multiprocessing` (procesos separados) y no `threading` (hilos):
en Python los hilos comunes no dan paralelismo real para trabajo de CPU,
por el GIL (Global Interpreter Lock) — solo un hilo ejecuta bytecode
Python a la vez. Los procesos de `multiprocessing` sí corren en paralelo
de verdad en distintos núcleos, que es lo que este curso pide demostrar.

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
python main.py \
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
proceso principal (productor + reductor)         pool de N procesos worker
─────────────────────────────────────────        ──────────────────────────
stream_csv() lee el CSV línea a línea    -----> process_row() valida,
(sin cargarlo entero a memoria) y reparte       transforma y calcula
las filas al pool vía imap_unordered            las variables derivadas
        │                                                │
        └──────────────── Result ───────────────────────┘
        │
        ▼
escribe el CSV limpio y acumula el resumen
(único lugar que toca ese estado -> sin locks)
```

- Cada worker recibe una fila y devuelve un `Result`, sin tocar nada fuera
  de su propio proceso — al ser procesos separados (no hilos ni memoria
  compartida), no hay condición de carrera posible entre workers, por
  construcción.
- Solo el proceso principal escribe el CSV de salida y acumula los
  contadores del resumen — es el único "reductor", así que tampoco hace
  falta ningún lock ahí.
- El número de workers es un flag (`--workers`) pensado para el benchmark
  de Speedup del Entregable 2 (probar 1, 2, 4, 8...).
- `stream_csv()` es un generador: el archivo se sigue leyendo línea por
  línea (nunca se carga entero en memoria), aunque las filas se agrupan en
  lotes (`chunksize`) antes de repartirlas al pool para no perder
  rendimiento serializando fila por fila.

## Estructura del proyecto

| Módulo | Archivos | Quién lo desarrolla |
|---|---|---|
| Esquema + lector concurrente (productor) | `cleaning/schema.py`, `cleaning/reader.py` | Loana (base) |
| Validación y transformación (worker) | `cleaning/worker.py` | Por definir (Rodrigo o Eduardo) |
| Orquestación, resumen y CLI | `cleaning/pipeline.py`, `cleaning/summary.py`, `cleaning/testdata_gen.py`, `main.py` | Por definir (Rodrigo o Eduardo) |

## Pruebas

Se incluye un generador de CSV sintético con casos borde
(`cleaning/testdata_gen.py`) usado para validar el pipeline sin depender
del archivo real de 470 MB:

```bash
python -m cleaning.testdata_gen data/testdata/clean_input.csv
python main.py --input data/testdata/clean_input.csv \
  --output data/testdata/clean_output.csv \
  --summary data/testdata/resumen.json --workers 4
```

Cubre: fila vacía, número de columnas incorrecto, error de parseo, `type`
inválido, monto negativo y saldo negativo (13 filas leídas, 7 válidas, 6
descartadas — una de cada motivo). Ya se corrió con 1, 2, 4 y 8 workers y
da exactamente el mismo resultado en todos los casos.
