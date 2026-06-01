"""Load knowledge-base files (.txt, .pdf, .docx) into LangChain documents."""

import os
from typing import List

from langchain_community.document_loaders import Docx2txtLoader, PyPDFLoader, TextLoader
from langchain_core.documents import Document

SUPPORTED_EXTENSIONS = (".txt", ".pdf", ".docx")

_EXT_LOADERS = {
    ".txt": lambda path: TextLoader(path, encoding="utf-8"),
    ".pdf": PyPDFLoader,
    ".docx": Docx2txtLoader,
}


def supported_extensions() -> tuple[str, ...]:
    return SUPPORTED_EXTENSIONS


def is_supported_filename(filename: str) -> bool:
    return os.path.splitext(filename)[1].lower() in SUPPORTED_EXTENSIONS


def load_file(domain_id: str, file_path: str) -> List[Document]:
    """Load a single knowledge-base file and attach domain metadata."""
    filename = os.path.basename(file_path)
    ext = os.path.splitext(filename)[1].lower()
    loader_factory = _EXT_LOADERS.get(ext)
    if loader_factory is None:
        raise ValueError(f"Unsupported file type: {ext or filename}")

    print(f"Loading [{domain_id}] {filename}")
    loader = loader_factory(file_path)
    docs = loader.load()
    for doc in docs:
        if doc.metadata is None:
            doc.metadata = {}
        doc.metadata["filename"] = filename
        doc.metadata["domain_id"] = domain_id
        doc.metadata["source_file"] = filename
        doc.metadata["file_type"] = ext.lstrip(".")
    return docs
