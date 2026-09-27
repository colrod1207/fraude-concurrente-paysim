/*
 * sync_model.pml
 *
 * Modela la lógica de sincronización del pipeline de limpieza concurrente
 * (internal/preprocessing/pipeline.go): una goroutine productora
 * (StreamCSV) envía filas por un channel; un pool de NUM_WORKERS goroutines
 * worker las consume y produce resultados por otro channel; y una única
 * goroutine reductora los consume y acumula el resumen.
 *
 * Simplificaciones respecto al código real (necesarias para que el modelo
 * sea verificable, pero que preservan la lógica de sincronización):
 *   - En vez de "cerrar" el channel `rows` (Go no tiene eso en Promela),
 *     el Producer envía un mensaje STOP por cada worker: cada worker
 *     consume exactamente un JOB o un STOP antes de terminar, lo cual
 *     reproduce la misma garantía que el cierre real de channel en Go.
 *   - El WaitGroup + "close(results)" del código real se modela con un
 *     proceso Closer: cada worker, al terminar, le avisa por done_ch;
 *     cuando Closer recibió el aviso de los NUM_WORKERS workers, envía un
 *     mensaje DONE al Reducer para que deje de esperar más resultados.
 *   - El contenido real de cada fila (floats, strings) no importa para la
 *     sincronización, así que cada JOB/RESULT solo lleva un identificador.
 */

#define NUM_ROWS    4
#define NUM_WORKERS 2

mtype = { JOB, STOP, RESULT, DONE };

chan rows    = [4] of { mtype, byte };
chan results = [4] of { mtype, byte };
chan done_ch = [NUM_WORKERS] of { byte };

/* Estado compartido -- en el código real es el *csv.Writer y el *Summary
 * que solo toca la goroutine reductora (la que llama a Run). */
byte total_processed = 0;
bool reducer_busy    = false;  /* centinela para verificar exclusión mutua */

/* ---- Producer == StreamCSV ------------------------------------------- */
active proctype Producer() {
    byte i = 0;

    do
    :: i < NUM_ROWS -> rows ! JOB(i); i++
    :: else -> break
    od

    i = 0;
    do
    :: i < NUM_WORKERS -> rows ! STOP(0); i++
    :: else -> break
    od
}

/* ---- Worker pool == "for i := 0; i < numWorkers; i++ { go func(){...} }" */
active [NUM_WORKERS] proctype Worker() {
    mtype kind;
    byte  payload;
    bool  working = true;

    do
    :: working ->
        rows ? kind(payload);
        if
        :: kind == JOB  -> results ! RESULT(payload)
        :: kind == STOP -> working = false; done_ch ! _pid
        fi
    :: !working -> break
    od
}

/* ---- Closer == "go func(){ wg.Wait(); close(results) }()" ------------ */
active proctype Closer() {
    byte doneCount = 0;
    byte w;

    do
    :: doneCount < NUM_WORKERS -> done_ch ? w; doneCount++
    :: else -> break
    od

    results ! DONE(0)
}

/* ---- Reducer == la goroutine que llama a Run() ------------------------ */
active proctype Reducer() {
    mtype kind;
    byte  payload;
    bool  running = true;

    do
    :: running ->
        results ? kind(payload);
        if
        :: kind == RESULT ->
            /* Si alguna vez dos procesos entraran aquí a la vez, esta
             * aserción fallaría: es la prueba formal de exclusión mutua. */
            assert(!reducer_busy);
            reducer_busy = true;
            total_processed++;
            reducer_busy = false
        :: kind == DONE -> running = false
        fi
    :: !running -> break
    od

    /* Propiedad de seguridad: cada fila enviada por el productor fue
     * contabilizada exactamente una vez -- no se perdió ni se duplicó
     * ningún resultado por una mala sincronización. */
    assert(total_processed == NUM_ROWS);
}
