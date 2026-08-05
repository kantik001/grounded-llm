"""Compatibility shim — prefer ``api.http.app``.

Gunicorn / tests may still use ``api.app:app``.
When run as ``python api/app.py``, bootstrap the repo root onto ``sys.path``
so ``import api`` resolves (script dir alone is ``api/``, not the monorepo root).
"""

from __future__ import annotations

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if _root not in sys.path:
    sys.path.insert(0, _root)

from api.http.app import app

__all__ = ["app"]

if __name__ == "__main__":
    port = int(os.environ.get("PYTHON_SERVICE_PORT", 5000))
    app.run(host="0.0.0.0", port=port, debug=False)
