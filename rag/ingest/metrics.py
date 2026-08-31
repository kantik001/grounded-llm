"""In-process counters for ingest workers (Prometheus-friendly text)."""

from __future__ import annotations

import threading
import time

_lock = threading.Lock()
_counters: dict[str, int] = {}
_durations: dict[str, list[float]] = {}


def inc(name: str, value: int = 1) -> None:
    with _lock:
        _counters[name] = _counters.get(name, 0) + value


def observe_duration(name: str, seconds: float) -> None:
    with _lock:
        bucket = _durations.setdefault(name, [])
        bucket.append(seconds)
        if len(bucket) > 200:
            del bucket[:100]


class timer:
    def __init__(self, name: str) -> None:
        self.name = name
        self._start = 0.0

    def __enter__(self):
        self._start = time.perf_counter()
        return self

    def __exit__(self, *args):
        observe_duration(self.name, time.perf_counter() - self._start)


def prometheus_lines() -> list[str]:
    lines: list[str] = []
    with _lock:
        counters = dict(_counters)
        durations = {k: list(v) for k, v in _durations.items()}
    for key, val in sorted(counters.items()):
        metric = key.replace(".", "_")
        lines.append(f"grounded_ingest_{metric} {val}")
    for key, samples in sorted(durations.items()):
        if not samples:
            continue
        metric = key.replace(".", "_")
        total = sum(samples)
        lines.append(f"grounded_ingest_{metric}_seconds_sum {total:.6f}")
        lines.append(f"grounded_ingest_{metric}_seconds_count {len(samples)}")
    return lines
