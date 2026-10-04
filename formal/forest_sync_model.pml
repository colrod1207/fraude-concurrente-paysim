/*
 * forest_sync_model.pml
 *
 * Modela la lógica de sincronización de TrainForestConcurrent
 * (internal/ml/forest_concurrent.go): los índices de árbol
 * se cargan en un channel "jobs"; un pool de NUM_WORKERS goroutines toma
 * índices del channel y entrena cada árbol, guardándolo en su propia
 * posición de forest.trees[i]. No hay channel de resultados ni reductor:
 * cada worker escribe directamente en su índice, y como cada índice lo
 * recibe un solo worker, no hace falta ningún lock.
 *
 * Diferencia clave con sync_model.pml: ahí varios
 * resultados confluyen en un reductor que los serializa; acá no hay
 * confluencia, cada resultado tiene su propia celda de destino. Por eso
 * se verifican propiedades distintas; en vez de "un solo proceso toca el
 * estado compartido", acá se verifica "cada celda del arreglo compartido
 * la escribe como máximo un proceso, y exactamente una vez".
 *
 * Simplificación respecto al código real: en Go, main() llena el channel
 * jobs por completo y lo cierra
 * antes de que importe si los workers ya empezaron a consumir. Acá se
 * modela con un proceso Loader que llena los JOB y, como Promela no tiene
 * "close", agrega un STOP por worker al final 
 * para representar el cierre del channel.
 *
 * Para saber cuándo los NUM_WORKERS workers terminaron (equivalente a
 * sync.WaitGroup.Wait()) se usa el mismo patrón de "Closer" que en
 * sync_model.pml: cada worker, al recibir STOP, avisa por done_ch antes de
 * terminar, y un proceso Checker cuenta esos avisos. (Una primera versión
 * de este archivo intentó usar la variable interna _nr_pr de Promela para
 * detectar "todos terminaron" sin un channel explícito; SPIN reportó un
 * invalid end state con esa versión porque _nr_pr no se comporta como un
 * contador simple de procesos vivos. Quedó como lección: en Promela, igual
 * que en Go, mejor señalizar explícitamente por un channel que depender de
 * contar procesos por fuera.)
 */

#define NUM_TREES   6
#define NUM_WORKERS 3

mtype = { JOB, STOP };

chan jobs    = [NUM_TREES + NUM_WORKERS] of { mtype, byte };
chan done_ch = [NUM_WORKERS] of { byte };

/* Estado compartido == forest.trees[i] en el código real.
 * 0 = arbol no entrenado todavia, 1 = entrenado. */
bit trees[NUM_TREES];

/* ---- Loader == "jobs := make(chan int, numTrees); for i := range seeds {
 * jobs <- i }; close(jobs)" en TrainForestConcurrent -------------------- */
active proctype Loader() {
    byte i = 0;

    do
    :: i < NUM_TREES -> jobs ! JOB(i); i++
    :: else -> break
    od

    i = 0;
    do
    :: i < NUM_WORKERS -> jobs ! STOP(0); i++
    :: else -> break
    od
}

/* ---- Worker pool == "for w := 0; w < numWorkers; w++ { go func(){ for i
 * := range jobs { forest.trees[i] = trainForestTree(...) } }() }" ------- */
active [NUM_WORKERS] proctype Worker() {
    mtype kind;
    byte  idx;
    bool  working = true;

    do
    :: working ->
        jobs ? kind(idx);
        if
        :: kind == JOB ->
            /* Exclusión mutua por construcción: cada índice de árbol lo
             * entrega el channel a un único worker. Si por algún bug dos
             * procesos recibieran el mismo índice, esta aserción fallaría
             * al segundo intento de escritura -- SPIN la exploraría en
             * todos los entrelazados posibles, no solo en el que a uno se
             * le ocurra probar a mano. */
            assert(trees[idx] == 0);
            trees[idx] = 1
        :: kind == STOP -> working = false; done_ch ! _pid
        fi
    :: !working -> break
    od
}

/* ---- Checker == "wg.Wait(); return forest, nil" ------------------------
 * Espera el aviso de los NUM_WORKERS workers (equivalente al WaitGroup) y
 * entonces confirma que los NUM_TREES árboles quedaron entrenados
 * exactamente una vez: ni se perdió ningún índice, ni se procesó dos
 * veces. */
active proctype Checker() {
    byte doneCount = 0;
    byte w;
    byte i;

    do
    :: doneCount < NUM_WORKERS -> done_ch ? w; doneCount++
    :: else -> break
    od

    i = 0;
    do
    :: i < NUM_TREES -> assert(trees[i] == 1); i++
    :: else -> break
    od
}
