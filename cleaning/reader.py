"""Lector concurrente (productor) del CSV de PaySim."""

import csv
from typing import Iterator, List, Optional


def stream_csv(path: str) -> Iterator[Optional[List[str]]]:
    with open(path, "r", encoding="utf-8", newline="") as f:
        reader = csv.reader(f)
        it = iter(reader)
        first = True
        while True:
            try:
                row = next(it)
            except StopIteration:
                break
            except csv.Error:
                yield None
                continue
            if first:
                first = False
                continue
            yield row
