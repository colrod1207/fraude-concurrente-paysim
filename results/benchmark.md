| Tarea | Modo | Workers | Media recortada (s) | Speedup | Eficiencia | CPU (%) | Núcleos usados | Heap pico (MB) | Memoria asignada (MB) |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| limpieza | secuencial | 1 | 14.4935 | 1.00 | 1.00 | 16.4 | 1.97 | 4.5 | 4752.2 |
| limpieza | concurrente | 1 | 13.4157 | 1.08 | 1.08 | 22.0 | 2.64 | 4.7 | 4752.5 |
| limpieza | concurrente | 2 | 14.0190 | 1.03 | 0.52 | 23.0 | 2.76 | 4.6 | 4752.7 |
| limpieza | concurrente | 4 | 14.5131 | 1.00 | 0.25 | 25.1 | 3.01 | 4.7 | 4752.9 |
| limpieza | concurrente | 8 | 14.8399 | 0.98 | 0.12 | 25.6 | 3.07 | 4.5 | 4753.0 |
| random-forest | secuencial | 1 | 12.9868 | 1.00 | 1.00 | 9.7 | 1.16 | 82.1 | 5545.2 |
| random-forest | concurrente | 1 | 13.0702 | 0.99 | 0.99 | 9.6 | 1.15 | 80.8 | 5545.2 |
| random-forest | concurrente | 2 | 8.6792 | 1.50 | 0.75 | 20.3 | 2.44 | 107.0 | 5545.2 |
| random-forest | concurrente | 4 | 6.6018 | 1.97 | 0.49 | 38.0 | 4.55 | 173.2 | 5545.2 |
| random-forest | concurrente | 8 | 6.0680 | 2.14 | 0.27 | 63.8 | 7.66 | 264.0 | 5545.2 |
