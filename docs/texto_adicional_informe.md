# Texto adicional para el informe

Dos bloques nuevos para pegar en el informe, además de las conclusiones y
recomendaciones de `conclusiones_recomendaciones.md`.

## Párrafo para la sección 5 (Promela), después del modelo de limpieza

Además del modelo de la limpieza, se modeló en Promela el patrón del Random
Forest concurrente (`formal/forest_sync_model.pml`), que tiene una topología
distinta: no hay channel de resultados ni reductor. Un proceso Loader carga
los índices de árbol en un channel `jobs` y un pool de 3 Workers toma
índices y marca cada posición `trees[i]`. Se verificaron dos aserciones: que
ninguna posición se escriba dos veces (`assert(trees[idx] == 0)`, es decir,
exclusión mutua por índice) y que al final las 6 posiciones estén escritas
exactamente una vez. SPIN 6.5.2 exploró 106,330 estados (167,023 transiciones,
profundidad 110) con 0 errores y sin estados inalcanzables en Loader ni Worker.
La evidencia está en `formal/evidencia_forest_spin.txt`. Esto respalda
formalmente por qué `TrainForestConcurrent` no necesita locks.

## Sección nueva: Análisis de código asistido por IA

Se pidió a Claude (Anthropic, vía Claude.ai) una revisión del repositorio
con un prompt estructurado: rol de revisor senior de Go y concurrencia,
ejecución real de `go build`, `go vet`, `gofmt` y `go test -race -cover`,
clasificación de los hallazgos (GAPs) en calidad de código, seguridad,
patrones de concurrencia y otros, con ubicación, severidad y recomendación
accionable, y la instrucción de no inventar problemas. El prompt completo y
los resultados están en `docs/ai_code_review.md`.

Resultado (sobre `main` aae0c83): build, vet y gofmt sin errores, tests con
`-race` en verde y cobertura de 69 % a 99 % en los paquetes con lógica. Se
reportaron 11 GAPs, todos de severidad media o baja, y ninguna condición de
carrera ni bloqueo.

Tabla 12. Resumen de hallazgos y acciones del equipo

| Categoría | GAPs reales | Severidad máxima | Acción del equipo |
|---|---:|---|---|
| Calidad de código | 4 (C1 a C4) | Media | Pendiente (trabajo futuro) |
| Seguridad | 2 (S1, S3) más 1 buena práctica (S2) | Media | Pendiente |
| Concurrencia | 2 (P1, P2) | Baja | P1 resuelto con el Promela del forest |
| Otros | 3 (O1 a O3) | Media | Resueltos: `go 1.22`, CI con GitHub Actions y licencia MIT |

Los pendientes más relevantes son GAP-C1 (el error de parseo CSV se cuenta
como `empty_row`), GAP-S1 (no se valida un SHA-256 del `.zip` descargado) y
GAP-C4 (la lógica de `main` no tiene tests). El detalle de cada hallazgo y su
seguimiento están en `docs/ai_code_review.md`.
