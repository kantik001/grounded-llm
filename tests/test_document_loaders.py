"""Tests for rag/document_loaders.py."""

import os

import pytest

from rag.document_loaders import is_supported_filename, load_file, supported_extensions


def test_supported_extensions():
    assert ".txt" in supported_extensions()
    assert ".pdf" in supported_extensions()
    assert ".docx" in supported_extensions()


def test_is_supported_filename():
    assert is_supported_filename("policy.txt")
    assert is_supported_filename("guide.PDF")
    assert is_supported_filename("manual.docx")
    assert not is_supported_filename("notes.md")
    assert not is_supported_filename("archive.zip")


def test_load_txt_file():
    root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    path = os.path.join(root, "data", "default", "policy_vacation.txt")
    if not os.path.isfile(path):
        pytest.skip("demo policy_vacation.txt not found")

    docs = load_file("default", path)
    assert len(docs) >= 1
    assert docs[0].metadata["filename"] == "policy_vacation.txt"
    assert docs[0].metadata["domain_id"] == "default"
    assert docs[0].metadata["file_type"] == "txt"
    assert "отпуск" in docs[0].page_content.lower()


def test_unsupported_extension_raises():
    with pytest.raises(ValueError, match="Unsupported"):
        load_file("default", "/tmp/readme.md")
