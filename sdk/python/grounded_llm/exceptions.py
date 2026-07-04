"""Grounded LLM API errors."""


class GroundedAPIError(Exception):
    """Raised when the API returns an error response."""

    def __init__(self, message: str, status_code: int | None = None, payload: dict | None = None):
        super().__init__(message)
        self.status_code = status_code
        self.payload = payload or {}


class GroundedAuthError(GroundedAPIError):
    """Raised on 401/403 responses."""
