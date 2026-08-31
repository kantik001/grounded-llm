"""Postgres connection helpers for KB modules."""

from __future__ import annotations

import os
from contextlib import contextmanager


def database_url() -> str:
    url = (os.environ.get("DATABASE_URL") or "").strip()
    if not url:
        raise RuntimeError("DATABASE_URL is required for KB document registry")
    return url


def psycopg_dsn(connection: str) -> str:
    return connection.replace("postgresql+psycopg://", "postgresql://", 1)


@contextmanager
def connect():
    import psycopg

    with psycopg.connect(psycopg_dsn(database_url())) as conn:
        yield conn
