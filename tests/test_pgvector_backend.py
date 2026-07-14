"""Tests for pgvector vector backend."""

import os

import pytest
from rag.vector_backend import get_vector_backend, reset_vector_backend
from rag.vector_backend.pgvector_backend import (
    normalize_pg_connection,
    pg_connection_url,
    psycopg_dsn,
)


@pytest.fixture(autouse=True)
def _reset_backend():
    reset_vector_backend()
    yield
    reset_vector_backend()


def test_normalize_pg_connection_postgres_scheme():
    assert (
        normalize_pg_connection("postgres://user:pass@localhost:5432/db")
        == "postgresql+psycopg://user:pass@localhost:5432/db"
    )


def test_normalize_pg_connection_postgresql_scheme():
    assert (
        normalize_pg_connection("postgresql://user:pass@localhost:5432/db")
        == "postgresql+psycopg://user:pass@localhost:5432/db"
    )


def test_normalize_pg_connection_already_psycopg():
    url = "postgresql+psycopg://user:pass@localhost:5432/db"
    assert normalize_pg_connection(url) == url


def test_psycopg_dsn_strips_driver():
    assert (
        psycopg_dsn("postgresql+psycopg://user:pass@localhost:5432/db")
        == "postgresql://user:pass@localhost:5432/db"
    )


def test_pg_connection_url_prefers_pgvector_url():
    os.environ["PGVECTOR_URL"] = "postgres://pg@db:5432/rag"
    os.environ["DATABASE_URL"] = "postgres://app@db:5432/app"
    try:
        assert pg_connection_url() == "postgresql+psycopg://pg@db:5432/rag"
    finally:
        os.environ.pop("PGVECTOR_URL", None)
        os.environ.pop("DATABASE_URL", None)


def test_pg_connection_url_missing_raises():
    os.environ.pop("PGVECTOR_URL", None)
    os.environ.pop("DATABASE_URL", None)
    with pytest.raises(RuntimeError, match="PGVECTOR_URL or DATABASE_URL"):
        pg_connection_url()


def test_pgvector_backend_factory(monkeypatch):
    class FakePGVectorBackend:
        def reset(self) -> None:
            pass

    monkeypatch.setattr(
        "rag.vector_backend.pgvector_backend.PGVectorBackend",
        FakePGVectorBackend,
    )
    os.environ["VECTOR_STORE"] = "pgvector"
    try:
        backend = get_vector_backend()
        assert backend.__class__.__name__ == "FakePGVectorBackend"
    finally:
        os.environ.pop("VECTOR_STORE", None)
