# Contenido para Gamma (presentación de 6 minutos)

**Instrucciones para Gamma:** crea una presentación de **9 diapositivas** en
español, tono académico y sobrio, sin imágenes de stock. Usa tablas y
diagramas simples en vez de párrafos largos. Cada diapositiva debe tener un
título corto, un máximo de 5 viñetas de pocas palabras y, donde se indica, la
tabla o el diagrama. Fondo claro, una sola paleta (azul oscuro y un acento
naranja). Pon las notas del orador en las notas de cada diapositiva.

Contexto: curso CC65 Programación Concurrente y Distribuida (UPC), Trabajo
Parcial, Grupo 10. Integrantes: Rodrigo Alejandro Meza Polo, Eduardo Fernando
Bravo Lévano y Loana Colleen Rodríguez Matos. Repositorio:
https://github.com/colrod1207/fraude-concurrente-paysim

---

## Diapositiva 1: Portada
- Título: **Detección de fraude concurrente en Go con PaySim**
- Subtítulo: Informe de la PC2, Grupo 10
- Curso CC65, UPC; los tres nombres de los integrantes
- Frase: 6.3 millones de transacciones, secuencial vs. concurrente

*Notas:* presentarse y decir el objetivo en una frase: acelerar con goroutines
la limpieza de datos y el entrenamiento de un Random Forest, y medir cuánto se
gana.

## Diapositiva 2: Problema y datos
- Fraude en dinero móvil: pérdidas y falta de confianza (ODS 16, meta 16.5)
- Dataset PaySim: 6,362,620 transacciones, 11 columnas, 30 días simulados
- Solo 8,213 fraudes (0.13 %): clases muy desbalanceadas
- El fraude solo ocurre en TRANSFER y CASH_OUT
- Solo librería estándar de Go, sin librerías de ML

*Notas:* explicar que el desbalance obliga a mirar el recall y no solo el
accuracy.

## Diapositiva 3: Arquitectura
Diagrama de dos flujos:
- **Limpieza:** Productor (lee CSV) → channel → N Workers (validan y
  transforman) → channel → Reductor (escribe CSV y resumen)
- **Random Forest:** channel de índices de árbol → N Workers (entrenan un árbol
  cada uno) → cada uno guarda en su posición `trees[i]`
- Mensaje clave: **sin ningún mutex**; la exclusión mutua es por construcción

*Notas:* patrones usados: pipeline, worker pool, productor/consumidor/reductor.

## Diapositiva 4: El repositorio
Lista corta de carpetas:
- `internal/preprocessing`: limpieza
- `internal/ml`: Random Forest secuencial y concurrente
- `internal/benchmark` y `cmd/benchmark`: medición
- `formal/`: modelos Promela y evidencia de SPIN
- `results/` y `docs/`: resultados, informe de IA, conclusiones
- Un solo comando: `go run ./cmd/benchmark`

*Notas:* en la grabación, mostrar el repo en GitHub, las ramas feature, los
Pull Requests y el CI en verde.

## Diapositiva 5: Resultados del Random Forest
Tabla (32 árboles, profundidad 8, 160,000 muestras):

| Versión | Media recortada (s) | Speedup | Eficiencia |
|---|---:|---:|---:|
| Secuencial | 12.99 | 1.00 | 1.00 |
| Concurrente, 1 worker | 13.07 | 0.99 | 0.99 |
| Concurrente, 2 workers | 8.68 | 1.50 | 0.75 |
| Concurrente, 4 workers | 6.60 | 1.97 | 0.49 |
| Concurrente, 8 workers | 6.07 | 2.14 | 0.27 |

- Escala porque cada árbol es CPU intensivo e independiente
- La eficiencia baja: memoria compartida, 4 de 8 núcleos son de eficiencia

*Notas:* 10 corridas por configuración, media recortada del 10 %. Equipo: Intel
Core i5-13420H, Windows 11, Go 1.27.

## Diapositiva 6: Resultados de la limpieza
Tabla (6,362,620 filas):

| Versión | Media recortada (s) | Speedup |
|---|---:|---:|
| Secuencial | 14.49 | 1.00 |
| Concurrente, 1 worker | 13.42 | 1.08 |
| Concurrente, 2 workers | 14.02 | 1.03 |
| Concurrente, 4 workers | 14.51 | 1.00 |
| Concurrente, 8 workers | 14.84 | 0.98 |

- **No escala:** la lectura y la escritura son secuenciales (ley de Amdahl)
- Heap pico de ~5 MB gracias al streaming
- Mejora propuesta: enviar filas en bloques (batching)

*Notas:* contrastar con la diapositiva anterior: misma herramienta, resultado
distinto, porque cambia dónde está el cuello de botella.

## Diapositiva 7: Calidad del modelo
Matriz de confusión (40,001 muestras de prueba):

| | Predicho fraude | Predicho normal |
|---|---:|---:|
| Real fraude | 29 | 1 |
| Real normal | 0 | 39,971 |

- Precision 1.0000, recall 0.9667, F1 0.9831
- 29 de 30 fraudes detectados, 0 falsas alarmas
- Limitación: solo 30 fraudes en test, el recall varía mucho

## Diapositiva 8: Verificación y análisis con IA
- **Promela + SPIN:** dos modelos (limpieza y Random Forest), 0 errores
  (36,372 y 106,330 estados)
- `go test -race` sin carreras y CI en GitHub Actions
- **Revisión con Claude:** prompt estructurado, 11 GAPs de severidad media o
  baja, ninguna condición de carrera
- Resueltos: Promela del forest, `go 1.22`, CI, licencia MIT
- Pendientes: parse error CSV, SHA-256 del zip, tests de `main`

*Notas:* decir que el prompt y los hallazgos están en `docs/ai_code_review.md`.

## Diapositiva 9: Conclusiones y recomendaciones
Conclusiones:
- Concurrencia ayuda donde el trabajo es pesado e independiente (Random Forest, 2.14x)
- Donde domina la E/S, Amdahl limita el speedup (limpieza, ~1x)
- Diseño sin locks, verificado formalmente

Recomendaciones:
- Batching en la limpieza y lectura en paralelo
- Reducir la memoria por worker en el Random Forest
- Evaluar con más fraudes y en otro hardware

Cierre: enlace al repositorio.
