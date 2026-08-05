"""Process-local Prometheus metrics for RAG retrieve (HTTP + gRPC).

Counters are labeled by ``protocol`` (http|grpc) and ``outcome``
(ok|business|error). Latency is a histogram per protocol.

Series live in the current OS process only: with ``gunicorn -w N`` each
worker has its own totals (scrape all workers or run with ``-w 1`` for a
single series). No ``prometheus_client`` multiprocess dependency.
"""

from __future__ import annotations

import threading

# Upper bounds in seconds (inclusive); +Inf is implicit.
_DURATION_BUCKETS: tuple[float, ...] = (
    0.005,
    0.01,
    0.025,
    0.05,
    0.1,
    0.25,
    0.5,
    1.0,
    2.5,
    5.0,
    10.0,
)

_PROTOCOLS = frozenset({"http", "grpc"})

_lock = threading.Lock()
# (protocol, outcome) -> count
_requests: dict[tuple[str, str], int] = {}
# protocol -> bucket counts (len = len(_DURATION_BUCKETS) + 1 for +Inf)
_hist_buckets: dict[str, list[int]] = {}
_hist_sum: dict[str, float] = {}
_hist_count: dict[str, int] = {}


def _normalize_protocol(protocol: str) -> str:
    p = (protocol or "http").strip().lower()
    return p if p in _PROTOCOLS else "http"


def _normalize_outcome(*, ok: bool, business_failure: bool) -> str:
    if ok:
        return "ok"
    if business_failure:
        return "business"
    return "error"


def record_retrieve(
    duration_seconds: float,
    *,
    protocol: str,
    ok: bool,
    business_failure: bool = False,
) -> None:
    """Record one retrieve call.

    ``ok=True`` — retrieval succeeded.
    ``ok=False, business_failure=True`` — handled miss / validation (no exception).
    ``ok=False, business_failure=False`` — unexpected exception / transport failure.
    """
    if duration_seconds < 0:
        duration_seconds = 0.0
    proto = _normalize_protocol(protocol)
    outcome = _normalize_outcome(ok=ok, business_failure=business_failure)

    with _lock:
        key = (proto, outcome)
        _requests[key] = _requests.get(key, 0) + 1

        buckets = _hist_buckets.get(proto)
        if buckets is None:
            buckets = [0] * (len(_DURATION_BUCKETS) + 1)
            _hist_buckets[proto] = buckets
            _hist_sum[proto] = 0.0
            _hist_count[proto] = 0

        placed = False
        for i, bound in enumerate(_DURATION_BUCKETS):
            if duration_seconds <= bound:
                buckets[i] += 1
                placed = True
                break
        if not placed:
            buckets[-1] += 1

        _hist_sum[proto] = _hist_sum.get(proto, 0.0) + duration_seconds
        _hist_count[proto] = _hist_count.get(proto, 0) + 1


def retrieve_metrics_lines() -> list[str]:
    with _lock:
        req_snapshot = {k: v for k, v in _requests.items()}
        hist_buckets = {p: list(b) for p, b in _hist_buckets.items()}
        hist_sum = dict(_hist_sum)
        hist_count = dict(_hist_count)

    lines: list[str] = [
        "# HELP rag_retrieve_requests_total RAG retrieve calls by protocol and outcome",
        "# TYPE rag_retrieve_requests_total counter",
    ]
    # Emit stable zero series for common labels so dashboards do not flap.
    for proto in ("http", "grpc"):
        for outcome in ("ok", "business", "error"):
            n = req_snapshot.get((proto, outcome), 0)
            lines.append(
                f'rag_retrieve_requests_total{{protocol="{proto}",outcome="{outcome}"}} {n}'
            )

    lines.extend(
        [
            "# HELP rag_retrieve_duration_seconds RAG retrieve latency histogram by protocol",
            "# TYPE rag_retrieve_duration_seconds histogram",
        ]
    )
    for proto in ("http", "grpc"):
        buckets = hist_buckets.get(proto) or [0] * (len(_DURATION_BUCKETS) + 1)
        cumulative = 0
        for i, bound in enumerate(_DURATION_BUCKETS):
            cumulative += buckets[i]
            lines.append(
                f'rag_retrieve_duration_seconds_bucket{{protocol="{proto}",le="{bound}"}} {cumulative}'
            )
        cumulative += buckets[-1]
        lines.append(
            f'rag_retrieve_duration_seconds_bucket{{protocol="{proto}",le="+Inf"}} {cumulative}'
        )
        lines.append(
            f'rag_retrieve_duration_seconds_sum{{protocol="{proto}"}} {hist_sum.get(proto, 0.0):.6f}'
        )
        lines.append(
            f'rag_retrieve_duration_seconds_count{{protocol="{proto}"}} {hist_count.get(proto, 0)}'
        )

    return lines


def reset_retrieve_metrics_for_tests() -> None:
    with _lock:
        _requests.clear()
        _hist_buckets.clear()
        _hist_sum.clear()
        _hist_count.clear()
