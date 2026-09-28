## Métricas del Random Forest (conjunto de test)

200000 muestras (159999 train / 40001 test, split 80/20 estratificado), 32 árboles, profundidad 8, semilla 42.

| | Predicho fraude | Predicho normal |
|---|---:|---:|
| **Real fraude** | TP = 29 | FN = 1 |
| **Real normal** | FP = 0 | TN = 39971 |

| Accuracy | Precision | Recall | F1 |
|---:|---:|---:|---:|
| 1.0000 | 1.0000 | 0.9667 | 0.9831 |
