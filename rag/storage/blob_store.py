"""Object storage for KB document blobs (local filesystem or S3-compatible)."""

from __future__ import annotations

import os
from pathlib import Path
from typing import Protocol


class BlobStore(Protocol):
    def put(self, key: str, data: bytes, *, content_type: str = "application/octet-stream") -> str: ...
    def get(self, key: str) -> bytes: ...
    def delete(self, key: str) -> None: ...
    def build_key(
        self,
        tenant_id: str,
        domain_id: str,
        document_id: str,
        version: int,
        sha256_hex: str,
        ext: str,
    ) -> str: ...


def build_versioned_key(
    tenant_id: str,
    domain_id: str,
    document_id: str,
    version: int,
    sha256_hex: str,
    ext: str,
) -> str:
    ext = ext.lstrip(".")
    suffix = f".{ext}" if ext else ""
    return f"tenants/{tenant_id}/domains/{domain_id}/docs/{document_id}/v{version}/{sha256_hex}{suffix}"


class LocalBlobStore:
    def __init__(self, root: str) -> None:
        self.root = Path(root)

    def build_key(
        self,
        tenant_id: str,
        domain_id: str,
        document_id: str,
        version: int,
        sha256_hex: str,
        ext: str,
    ) -> str:
        return build_versioned_key(tenant_id, domain_id, document_id, version, sha256_hex, ext)

    def _path(self, key: str) -> Path:
        clean = key.replace("\\", "/").lstrip("/")
        return self.root / clean

    def put(self, key: str, data: bytes, *, content_type: str = "application/octet-stream") -> str:
        path = self._path(key)
        path.parent.mkdir(parents=True, exist_ok=True)
        tmp = path.with_suffix(path.suffix + ".tmp")
        tmp.write_bytes(data)
        tmp.replace(path)
        return key

    def get(self, key: str) -> bytes:
        return self._path(key).read_bytes()

    def delete(self, key: str) -> None:
        path = self._path(key)
        if path.is_file():
            path.unlink()


class S3BlobStore:
    def __init__(
        self,
        *,
        endpoint: str,
        access_key: str,
        secret_key: str,
        bucket: str,
        prefix: str = "",
        use_ssl: bool = False,
        region: str = "us-east-1",
    ) -> None:
        from minio import Minio

        self.bucket = bucket
        self.prefix = prefix.strip("/")
        self.client = Minio(
            endpoint,
            access_key=access_key,
            secret_key=secret_key,
            secure=use_ssl,
            region=region,
        )

    def build_key(
        self,
        tenant_id: str,
        domain_id: str,
        document_id: str,
        version: int,
        sha256_hex: str,
        ext: str,
    ) -> str:
        return build_versioned_key(tenant_id, domain_id, document_id, version, sha256_hex, ext)

    def _full_key(self, key: str) -> str:
        key = key.replace("\\", "/").lstrip("/")
        if self.prefix:
            return f"{self.prefix}/{key}"
        return key

    def put(self, key: str, data: bytes, *, content_type: str = "application/octet-stream") -> str:
        import io

        self.client.put_object(
            self.bucket,
            self._full_key(key),
            io.BytesIO(data),
            length=len(data),
            content_type=content_type,
        )
        return key

    def get(self, key: str) -> bytes:
        resp = self.client.get_object(self.bucket, self._full_key(key))
        try:
            return resp.read()
        finally:
            resp.close()
            resp.release_conn()

    def delete(self, key: str) -> None:
        self.client.remove_object(self.bucket, self._full_key(key))


_store: BlobStore | None = None


def get_blob_store() -> BlobStore:
    global _store
    if _store is not None:
        return _store

    backend = (os.environ.get("KB_BLOB_BACKEND") or "local").strip().lower()
    if backend == "s3":
        endpoint = (os.environ.get("KB_S3_ENDPOINT") or "").strip()
        access = (os.environ.get("KB_S3_ACCESS_KEY") or "").strip()
        secret = (os.environ.get("KB_S3_SECRET_KEY") or "").strip()
        bucket = (os.environ.get("KB_S3_BUCKET") or "grounded-kb").strip()
        prefix = (os.environ.get("KB_S3_PREFIX") or "").strip()
        use_ssl = (os.environ.get("KB_S3_USE_SSL") or "").lower() in ("1", "true", "yes")
        region = (os.environ.get("KB_S3_REGION") or "us-east-1").strip()
        if not endpoint or not access or not secret:
            raise RuntimeError("KB_S3_ENDPOINT, KB_S3_ACCESS_KEY, KB_S3_SECRET_KEY required for s3 backend")
        _store = S3BlobStore(
            endpoint=endpoint,
            access_key=access,
            secret_key=secret,
            bucket=bucket,
            prefix=prefix,
            use_ssl=use_ssl,
            region=region,
        )
        return _store

    root = (os.environ.get("KB_BLOB_DIR") or "").strip()
    if not root:
        from rag.kb_discovery import data_dir

        root = os.path.join(data_dir(), "blobs")
    _store = LocalBlobStore(root)
    return _store


def reset_blob_store() -> None:
    global _store
    _store = None
