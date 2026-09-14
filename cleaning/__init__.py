"""Paquete de limpieza concurrente del dataset PaySim (PC1, CC65).

Implementa el patrón productor/consumidor/reductor descrito en el informe:
un generador lector (productor) entrega filas crudas del CSV, un pool de
procesos (multiprocessing) las valida y transforma en paralelo, y el
proceso principal actúa como único "reductor" que escribe el CSV limpio y
acumula el resumen -- así no hace falta ningún lock: nadie más comparte ese
estado.
"""
