# Conclusiones y recomendaciones

Texto listo para pegar en el informe. Todas las cifras salen de
`results/benchmark.md`, `results/metricas_random_forest.md` y
`docs/ai_code_review.md`.

## Conclusiones

1. **La concurrencia no ayuda por igual en todas las etapas.** El Random
   Forest alcanza un Speedup de 2.14 con 8 workers (eficiencia 0.27), porque
   entrenar cada árbol es trabajo de CPU independiente. La limpieza del
   dataset (6 362 620 filas) se queda en un Speedup de 0.98 a 1.08, sin
   importar cuántos workers se usen.
2. **La ley de Amdahl explica la limpieza.** Validar una fila cuesta muy poco
   frente a leerla, pasarla por dos channels y escribirla. El productor
   (lectura) y el reductor (escritura) son una sola goroutine cada uno, y esa
   parte secuencial limita el Speedup. Los núcleos usados se estancan en ~3.
3. **La eficiencia cae al subir los workers.** En el Random Forest pasa de
   0.75 con 2 workers a 0.27 con 8. Los workers compiten por memoria y caché
   (más de 5 GB asignados en total), y 4 de los 8 núcleos del equipo son de
   eficiencia y ~33 % más lentos.
4. **El diseño evita condiciones de carrera por construcción.** El proyecto no
   usa ningún `sync.Mutex`. En la limpieza solo la goroutine reductora escribe
   la salida y el resumen. En el Random Forest cada worker escribe en su propia
   posición `forest.trees[i]`. `go test -race` no reporta errores, y SPIN
   verificó ambos modelos Promela sin violaciones (`formal/evidencia_spin.txt`
   y `formal/evidencia_forest_spin.txt`).
5. **El modelo detecta el fraude con muy pocas falsas alarmas.** En el
   conjunto de test (40 001 muestras) obtuvo precision 1.0000, recall 0.9667 y
   F1 0.9831: 29 de 30 fraudes detectados y 0 falsas alarmas. Con solo 30
   fraudes en test, el recall tiene mucha varianza: un fraude más o menos
   cambia el resultado en ~3 puntos.
6. **La memoria se controla con streaming.** El heap pico de la limpieza es de
   ~5 MB porque el CSV nunca se carga entero. En el Random Forest crece de 82 a
   264 MB al pasar de 1 a 8 workers, porque cada worker guarda su muestra
   bootstrap.
7. **La revisión con IA aportó hallazgos útiles y verificables.** Claude
   analizó el repositorio con un prompt estructurado (calidad, seguridad,
   concurrencia, otros) y reportó 11 GAPs, ninguno con severidad alta. El
   equipo cerró GAP-P1, O1, O2 y O3. Las cuatro categorías incluyeron además puntos a favor que confirmaron
   buenas prácticas ya presentes, como la extracción del `.zip` sin *zip slip*.

## Recomendaciones

1. **Enviar las filas en bloques (batching) en la limpieza.** Reducir el costo
   por mensaje en los channels es la mejora con más probabilidad de acercar el
   Speedup a más de 1. Conviene medirla con el mismo benchmark.
2. **Paralelizar la lectura o la escritura del CSV.** Por ejemplo, partir el
   archivo en rangos de bytes y leerlos con varios productores. Es la única
   forma de reducir la parte secuencial que fija el límite de Amdahl.
3. **Reducir la presión de memoria del Random Forest.** Compartir los índices
   de la muestra bootstrap en vez de copiar filas, y reutilizar buffers de
   ordenamiento. Debería mejorar la eficiencia con 4 y 8 workers.
4. **Evaluar con más datos de fraude.** Con 30 fraudes en test, repetir el
   experimento con distintas semillas o validación cruzada estratificada daría
   un recall con intervalo de confianza en vez de un solo valor.
5. **Atender los GAPs pendientes del informe de IA.** En orden de prioridad:
   GAP-C1 (motivo propio `parse_error_csv`), GAP-S1 (SHA-256 del `.zip`
   descargado) y GAP-C4 (tests de la lógica de `main`).
6. **Medir en otro equipo.** Los resultados salen de una sola laptop con
   núcleos híbridos. Repetirlos en una máquina con núcleos homogéneos aislaría
   el efecto de los núcleos de eficiencia sobre la eficiencia reportada.
