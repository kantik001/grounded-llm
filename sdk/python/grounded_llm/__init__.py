"""Grounded LLM Python SDK — cited, verified document assistant API client."""

from grounded_llm.client import GroundedClient, MessageResult
from grounded_llm.exceptions import GroundedAPIError, GroundedAuthError

__all__ = ["GroundedClient", "MessageResult", "GroundedAPIError", "GroundedAuthError"]
__version__ = "0.2.0"
