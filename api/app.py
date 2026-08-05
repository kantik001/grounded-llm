"""Compatibility shim — prefer ``api.http.app``.

Gunicorn / tests may still use ``api.app:app``.
"""

from api.http.app import app

__all__ = ["app"]
