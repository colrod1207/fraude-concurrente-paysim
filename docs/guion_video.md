# Guion del video (6 minutos) y estructura del PPT

Reparto sugerido para 3 integrantes (A, B, C). Si son más o menos, se
reparten los bloques. Cada uno graba su parte por separado y se unen en Canva.
La primera diapositiva es la portada; el tiempo se cuenta desde la 2.

| Min | Quién | Diapositiva | Qué decir | Qué mostrar |
|---|---|---|---|---|
| 0:00–0:30 | A | 1. Portada y problema | Presentarse. Detectar fraude en PaySim (6.3 M transacciones) y comparar secuencial vs concurrente en Go. | Título, nombres, curso CC65. |
| 0:30–1:30 | A | 2. Arquitectura | Limpieza: productor → workers → reductor con goroutines y channels. Random Forest: worker pool, un árbol por job. No se usa ningún lock. | Diagrama del README. |
| 1:30–2:30 | A | 3. Repo en vivo | Recorrer la estructura: `internal/`, `cmd/`, `formal/`, `results/`, `docs/`. Mostrar `go run ./cmd/benchmark` como comando único. | Pantalla compartida del repo en GitHub. |
| 2:30–3:30 | B | 4. Resultados: Random Forest | Speedup 2.14 con 8 workers, eficiencia 0.27. Por qué escala (árboles independientes) y por qué cae la eficiencia (memoria y núcleos E). | Tabla de `results/benchmark.md`. |
| 3:30–4:15 | B | 5. Resultados: limpieza | Speedup ~1.0. Ley de Amdahl: leer y escribir son secuenciales. Heap de ~5 MB por streaming. | Tabla de limpieza. |
| 4:15–4:45 | B | 6. Calidad del modelo | Precision 1.0, recall 0.9667, F1 0.9831: 29 de 30 fraudes, 0 falsas alarmas. Mencionar que son pocos fraudes en test. | Matriz de confusión. |
| 4:45–5:15 | C | 7. Verificación formal | Dos modelos Promela verificados con SPIN, 0 errores. `go test -race` limpio y CI en GitHub Actions. | `formal/evidencia_forest_spin.txt`, badge o corrida de CI. |
| 5:15–5:45 | C | 8. Análisis con IA | Prompt estructurado a Claude, 11 hallazgos, ninguno de severidad alta. Cerramos 4 GAPs (Promela del forest, go.mod, CI, licencia). | `docs/ai_code_review.md`, sección 8. |
| 5:45–6:00 | C | 9. Conclusiones | Una frase por idea: concurrencia ayuda donde el trabajo es pesado e independiente; Amdahl limita el resto. Recomendación: batching. | Lista corta. |

## Consejos
- Ensayar con cronómetro: 6:00 es el límite, y cada parte está justa.
- No leer tablas completas. Decir solo la cifra clave de cada diapositiva.
- Grabar el repo en una ventana con letra grande (zoom del navegador al 125 %).
- Tener abierto de antemano `results/benchmark.md` y `docs/ai_code_review.md`.
