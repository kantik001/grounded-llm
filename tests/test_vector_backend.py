"""Tests for vector store backend factory."""

import os

import pytest
from rag.vector_backend import get_vector_backend, reset_vector_backend


@pytest.fixture(autouse=True)
def _reset_backend():
    reset_vector_backend()
    yield
    reset_vector_backend()


def test_default_backend_is_chroma():
    os.environ.pop("VECTOR_STORE", None)
    backend = get_vector_backend()
    assert backend.__class__.__name__ == "ChromaBackend"


def test_unknown_backend_raises():
    os.environ["VECTOR_STORE"] = "unknown_db"
    try:
        with pytest.raises(ValueError, match="Unknown VECTOR_STORE"):
            get_vector_backend()
    finally:
        os.environ.pop("VECTOR_STORE", None)
