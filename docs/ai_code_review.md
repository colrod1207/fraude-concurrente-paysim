# Informe de análisis de código con IA

**Repositorio analizado:** https://github.com/colrod1207/fraude-concurrente-paysim
**Commit analizado:** `aae0c83` (rama `main`)
**Modelo utilizado:** Claude (Anthropic), vía Claude.ai
**Fecha:** TP — CC65, Trabajo Parcial 2026-20

## 1. Prompt estructurado utilizado

```
Eres un revisor de código senior especializado en Go y en sistemas concurrentes.
Analiza el repositorio <URL>, commit <HASH>, con el siguiente alcance:

1. Clona el repositorio y lee todos los archivos .go (internal/, cmd/, main.go)
   y el modelo formal en formal/sync_model.pml.
2. Ejecuta, si el entorno lo permite: `go build ./...`, `go vet ./...`,
   `gofmt -l .`, `go test ./... -race -cover`. Reporta la salida real, no la
   infieras.
3. Identifica GAPs y clasifícalos en exactamente estas categorías:
   - Calidad de código (mantenibilidad, manejo de errores, duplicación,
     claridad, cobertura de tests)
   - Seguridad (descarga de datos externos, manejo de archivos, validación
     de entradas, dependencias)
   - Patrones de concurrencia (goroutines, channels, locks, condiciones de
     carrera, exclusión mutua, correctitud del paralelismo)
   - Otros (portabilidad, reproducibilidad, documentación, CI/CD)
4. Para cada GAP indica: ubicación exacta (archivo:línea o función),
   severidad (alta/media/baja), por qué es un problema, y una recomendación
   concreta y accionable.
5. No inventes hallazgos: si una sección del código está bien resuelta,
   dilo explícitamente en vez de forzar un problema.
6. Entrega el resultado en Markdown, listo para pegar en el repositorio.
```

## 2. Metodología

Se clonó el repositorio y se corrieron las siguientes verificaciones automáticas (el `go.mod` del proyecto declara `go 1.27.0`; el entorno de análisis solo tenía disponible Go 1.22.2, así que se usó una copia local del `go.mod` apuntando a 1.22.2 **solo para poder compilar y correr las herramientas** — el repositorio real no se modificó):

| Herramienta | Resultado |
|---|---|
| `go build ./...` | Sin errores |
| `go vet ./...` | Sin advertencias |
| `gofmt -l .` | Sin archivos mal formateados |
| `go test ./... -race` | Todos los paquetes con tests: **PASS** |
| `go test ./... -cover` | ver tabla abajo |

| Paquete | Cobertura |
|---|---:|
| internal/ml | 99.1% |
| internal/benchmark | 89.6% |
| internal/dataset | 76.2% |
| internal/preprocessing | 69.0% |
| main.go, cmd/* | 0.0% (sin tests — ver GAP-C4) |

Sobre esta base se hizo una revisión manual del código, el modelo Promela y la organización del repositorio.

## 3. Calidad de código

| ID | Ubicación | Severidad | Descripción | Recomendación |
|---|---|---|---|---|
| GAP-C1 | `internal/preprocessing/reader.go:36` | Media | Cuando `csv.Reader.Read()` devuelve un error de parseo (ej. comillas mal cerradas), `StreamCSV` envía `rows <- nil` en vez de propagar el error o marcar el motivo. Como `worker.go` llama a `isEmptyRow(nil)`, que devuelve `true`, esa fila termina contabilizada con el motivo `empty_row` en vez de un motivo propio como `parse_error_csv`. El `resumen_limpieza.json` queda técnicamente correcto en el conteo total, pero el desglose por motivo es engañoso para una fila real de error. | Agregar un motivo de descarte específico (p. ej. `CSVReadError`) y devolver un `Result{Discard: CSVReadError}` en vez de enviar `nil`. |
| GAP-C2 | `internal/ml/data.go:76` | Baja | `line, _ := reader.FieldPos(0)` se calcula en cada iteración del loop, incluso cuando la fila es válida y `line` no se usa. Es trabajo de más en una ruta que se ejecuta hasta 6+ millones de veces. | Calcularlo solo dentro de los `if err != nil` que efectivamente lo usan en el mensaje. |
| GAP-C3 | `internal/preprocessing/pipeline.go:103` | Baja | `formatAmount` usa `%.2f`, lo que trunca `errorBalanceOrig`/`errorBalanceDest` a 2 decimales. Si el dataset llegara a tener descuadres menores a 0.005, se redondearían a `0.00`, perdiendo precisión en una feature que el propio informe identifica como señal clave de fraude. En PaySim esto no ocurre en la práctica (los montos ya vienen con 2 decimales), pero es un supuesto implícito no documentado. | Documentar explícitamente en el comentario de la función por qué 2 decimales son suficientes para este dataset, o usar más precisión si se reutiliza el pipeline con otro dataset. |
| GAP-C4 | `main.go`, `cmd/benchmark`, `cmd/download`, `cmd/train` | Media | 0% de cobertura de tests en todos los `main`. Es razonable no testear `flag.Parse()`, pero funciones como `prepare()` en `main.go` o `parseWorkers()` en `cmd/benchmark/main.go` tienen lógica propia (parseo, validación de flags) que sí es testeable y hoy no tiene ningún test. | Extraer esa lógica a funciones puras ya separadas (como ya hicieron con `parseWorkers`) y agregarles tests unitarios, igual que al resto del proyecto. |
| — | General | — | **Punto a favor:** el resto del código (`internal/ml`, `internal/benchmark`, `internal/preprocessing`) tiene cobertura alta (69–99%), nombres de funciones y comentarios consistentes, y cada paquete explica su propósito en un comentario de cabecera. El manejo de errores con `%w` y mensajes contextualizados es correcto y uniforme en todo el repo. | — |

## 4. Seguridad

| ID | Ubicación | Severidad | Descripción | Recomendación |
|---|---|---|---|---|
| GAP-S1 | `internal/dataset/download.go` (`Download`) | Media | El `.zip` se descarga por HTTPS pero no se valida ninguna suma de verificación (hash) del contenido antes de usarlo como fuente de verdad del entrenamiento. Una descarga corrupta o interceptada no se detectaría; el pipeline simplemente entrenaría con datos silenciosamente alterados. | Agregar una verificación de SHA-256 del `.zip` (o del CSV extraído) contra un valor conocido, documentado en el README, antes de continuar. |
| GAP-S2 | `internal/dataset/download.go` (`extract`) | — (buena práctica) | La extracción del `.zip` **no** usa el nombre del archivo tal como viene en el entry (`file.Name`) como ruta destino — compara contra el `FileName` esperado y solo entonces escribe en `dest`, que es una ruta fija controlada por el programa. Esto evita por construcción un ataque de tipo *zip slip* (path traversal vía `../` en el nombre del archivo dentro del zip). | Ninguna; es el patrón correcto. Vale la pena dejarlo como comentario explícito en el código ("no usamos file.Name como ruta de salida, por seguridad"), para que quede claro que es una decisión consciente y no una casualidad. |
| GAP-S3 | `internal/dataset/download.go:14` | Baja | La URL de descarga (`KaggleURL`) es HTTPS pero está *hardcodeada* sin posibilidad de override salvo por flag en `cmd/download`; `EnsureRaw` (usado por `main.go` y `cmd/benchmark`) no expone esa opción. No es un hueco de seguridad explotable hoy, pero sí un acoplamiento rígido si Kaggle cambia su esquema de URLs. | Exponer la URL también como variable de entorno opcional, documentado, para no tener que tocar código si la URL cambia. |

## 5. Patrones de concurrencia

| ID | Ubicación | Severidad | Descripción |
|---|---|---|---|
| — | `internal/preprocessing/pipeline.go` (`Run`) | — (correcto) | Patrón productor/worker pool/reductor bien aplicado: el único proceso que escribe el CSV de salida y acumula el `Summary` es la goroutine que llama a `Run`; los workers son funciones puras sin estado compartido (confirmado también por el modelo Promela de PC2, verificado con SPIN, y por `go test -race` sin errores). |
| — | `internal/ml/forest_concurrent.go` | — (correcto) | Patrón distinto y bien justificado: en vez de un channel de resultados + reductor, cada worker escribe directamente en `forest.trees[i]`. Como los índices de un channel `jobs` sin duplicados garantizan que cada posición del slice la escribe un único worker, no hace falta ningún lock — y el comentario del código lo explica correctamente. Es una variación válida del patrón worker pool (fan-out sin fan-in explícito, porque no hay que combinar resultados, solo ensamblarlos por índice). |
| GAP-P1 | `internal/ml/forest_concurrent.go` (`TrainForestConcurrent`) | Baja | No hay verificación formal (Promela) específica para este patrón — el `.pml` existente de PC2 solo modela la limpieza (productor/worker/reductor con channel de resultados), que es una topología distinta a la de "workers escribiendo por índice en un array compartido". La corrección de este segundo patrón hoy descansa solo en `-race` y en el razonamiento del comentario, no en una verificación exhaustiva de estados. | Si da tiempo, agregar un segundo modelo `.pml` para este patrón específico (workers con índice único sobre un arreglo), análogo al que ya existe para la limpieza. No es estrictamente necesario porque el patrón es más simple de razonar que el productor/consumidor, pero sería más consistente con el rigor que ya mostraron en la limpieza. |
| GAP-P2 | `internal/benchmark/runner.go` (`ProfileRun`) | Baja | El goroutine muestreador de memoria (`stats.go`) usa `stop`/`wg.Wait()` para terminar limpiamente, que es correcto, pero no hay un límite de tiempo (timeout) si `fn()` se queda colgado — el `ProfileRun` esperaría indefinidamente. Dado que `fn` son llamadas a código propio y controlado (no I/O externo salvo la descarga, que si corresponde ya ocurrió antes), el riesgo real es bajo. | Opcional: envolver `fn()` con un `context.WithTimeout` si en el futuro se agregan pasos que dependan de red o de disco lento. |
| — | General | — (correcto) | No se encontraron locks (`sync.Mutex`) en ningún paquete del proyecto. Dado que ambos patrones de concurrencia usados (reductor único y escritura por índice exclusivo) garantizan exclusión mutua *por construcción* y no por sincronización explícita, la ausencia de locks es resultado de un diseño cuidadoso, no un descuido. |

## 6. Otros (portabilidad, reproducibilidad, documentación)

| ID | Ubicación | Severidad | Descripción | Recomendación |
|---|---|---|---|---|
| GAP-O1 | `go.mod:3` | Media | `go 1.27.0` es una versión muy reciente del lenguaje. Un evaluador o un compañero con un Go más antiguo instalado no podrá compilar el proyecto sin antes actualizar su toolchain (en este análisis, Go 1.22.2 no pudo compilarlo directamente). | A menos que el proyecto use específicamente una característica introducida en 1.27, bajar el requisito a una versión más ampliamente disponible (ej. 1.22 o 1.23), o documentar en el README la versión exacta de Go requerida y cómo instalarla. |
| GAP-O2 | raíz del repositorio | Baja | No hay carpeta `.github/workflows`: no hay integración continua que corra automáticamente `go build`, `go vet`, `gofmt -l` y `go test -race` en cada Pull Request. Hoy esas verificaciones dependen de que cada integrante las corra manualmente antes de hacer push. | Agregar un workflow simple de GitHub Actions (`go build ./...`, `go vet ./...`, `go test ./... -race`) que corra en cada PR — además de mejorar la calidad, deja evidencia automática y con sello de tiempo de que el código compilaba y pasaba los tests en cada entrega, útil frente a la regla de "no editar después de la fecha límite". |
| GAP-O3 | raíz del repositorio | Baja | No hay archivo `LICENSE`. El enunciado pide que el repositorio sea público, pero no especifica licencia; igual es una buena práctica para cualquier repo académico público. | Agregar una licencia simple (MIT, por ejemplo) si el equipo no tiene preferencia. |
| — | `README.md` | — (correcto) | El README explica con claridad cómo correr el proyecto con un solo comando, la arquitectura del pipeline con diagrama ASCII, y por qué no hace falta ningún lock — buena práctica de documentación que facilita justamente este tipo de revisión. |

## 7. Resumen

| Categoría | Hallazgos | Severidad más alta |
|---|---:|---|
| Calidad de código | 4 (3 reales + 1 de cobertura) | Media |
| Seguridad | 2 (1 real + 1 buena práctica confirmada) | Media |
| Patrones de concurrencia | 2 (ambos de bajo riesgo; el diseño base es correcto) | Baja |
| Otros | 3 | Media |

**No se encontraron condiciones de carrera, deadlocks ni violaciones de exclusión mutua** en el código Go (confirmado con `go test -race`) ni en el modelo formal verificado con SPIN. Los hallazgos de este informe son de severidad media/baja y apuntan sobre todo a robustez de bordes (manejo de errores de parseo), reproducibilidad (versión de Go) y proceso (ausencia de CI), no a errores funcionales del sistema.
