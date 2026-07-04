"""HTTP client for Grounded LLM REST API."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any, Iterator
from urllib.parse import urljoin

import requests

from grounded_llm.exceptions import GroundedAPIError, GroundedAuthError


@dataclass
class MessageResult:
    """Result of a chat message request."""

    session_id: str
    domain_id: str
    tenant_id: str
    messages: list[dict[str, Any]]
    success: bool = True
    error: str | None = None
    raw: dict[str, Any] = field(default_factory=dict)

    @property
    def last_assistant_message(self) -> dict[str, Any] | None:
        for msg in reversed(self.messages):
            if msg.get("role") == "assistant":
                return msg
        return None


class GroundedClient:
    """Client for Grounded LLM chat and public API endpoints."""

    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        *,
        api_key: str | None = None,
        tenant_id: str = "default",
        locale: str = "en",
        api_prefix: str = "/api/v1",
        timeout: float = 120.0,
        session: requests.Session | None = None,
    ):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.tenant_id = tenant_id
        self.locale = locale
        self.api_prefix = api_prefix.rstrip("/") or ""
        self.timeout = timeout
        self._session = session or requests.Session()

    def _headers(self, *, json_body: bool = True, stream: bool = False) -> dict[str, str]:
        headers: dict[str, str] = {
            "X-Tenant-ID": self.tenant_id,
            "X-Locale": self.locale,
        }
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        if json_body:
            headers["Content-Type"] = "application/json; charset=utf-8"
        if stream:
            headers["Accept"] = "text/event-stream"
        return headers

    def _url(self, path: str, *, versioned: bool = True) -> str:
        path = path if path.startswith("/") else f"/{path}"
        if versioned and self.api_prefix:
            if not path.startswith(self.api_prefix):
                path = f"{self.api_prefix}{path}"
        return urljoin(f"{self.base_url}/", path.lstrip("/"))

    def _request(
        self,
        method: str,
        path: str,
        *,
        versioned: bool = True,
        params: dict | None = None,
        json_body: dict | None = None,
        stream: bool = False,
    ) -> requests.Response:
        url = self._url(path, versioned=versioned)
        resp = self._session.request(
            method,
            url,
            headers=self._headers(json_body=json_body is not None, stream=stream),
            params=params,
            json=json_body,
            timeout=self.timeout,
            stream=stream,
        )
        if resp.status_code in (401, 403):
            raise GroundedAuthError(
                f"Authentication failed: HTTP {resp.status_code}",
                status_code=resp.status_code,
            )
        return resp

    def _parse_json(self, resp: requests.Response) -> dict[str, Any]:
        try:
            data = resp.json()
        except json.JSONDecodeError as exc:
            raise GroundedAPIError(
                f"Invalid JSON response: HTTP {resp.status_code}",
                status_code=resp.status_code,
            ) from exc
        if resp.status_code >= 400 or data.get("success") is False:
            msg = data.get("error") or f"HTTP {resp.status_code}"
            raise GroundedAPIError(str(msg), status_code=resp.status_code, payload=data)
        return data

    def health(self) -> dict[str, Any]:
        resp = self._request("GET", "/health", versioned=False)
        return self._parse_json(resp)

    def list_domains(self) -> dict[str, Any]:
        resp = self._request("GET", "/api/domains", versioned=False, params={"locale": self.locale})
        return self._parse_json(resp)

    def branding(self) -> dict[str, Any]:
        resp = self._request("GET", "/api/branding", versioned=False, params={"locale": self.locale})
        return self._parse_json(resp)

    def onboarding(self, domain_id: str = "default") -> dict[str, Any]:
        resp = self._request(
            "GET",
            "/api/onboarding",
            versioned=False,
            params={"domain_id": domain_id, "locale": self.locale},
        )
        return self._parse_json(resp)

    def create_session(self, domain_id: str = "default") -> str:
        data = self._parse_json(
            self._request("POST", "/session", json_body={"domain_id": domain_id})
        )
        session_id = data.get("session_id")
        if not session_id:
            raise GroundedAPIError("Missing session_id in response", payload=data)
        return str(session_id)

    def history(self, session_id: str) -> list[dict[str, Any]]:
        data = self._parse_json(
            self._request("GET", "/history", params={"session_id": session_id})
        )
        return list(data.get("messages") or [])

    def send_message(
        self,
        text: str,
        *,
        session_id: str,
        domain_id: str = "default",
        stream: bool = False,
    ) -> MessageResult:
        if stream:
            return self._send_message_stream(text, session_id=session_id, domain_id=domain_id)
        body = {
            "session_id": session_id,
            "domain_id": domain_id,
            "text": text,
        }
        data = self._parse_json(self._request("POST", "/message", json_body=body))
        return MessageResult(
            session_id=str(data.get("session_id", session_id)),
            domain_id=str(data.get("domain_id", domain_id)),
            tenant_id=str(data.get("tenant_id", self.tenant_id)),
            messages=list(data.get("messages") or []),
            success=True,
            raw=data,
        )

    def _iter_sse_events(self, resp: requests.Response) -> Iterator[tuple[str, dict[str, Any]]]:
        event_name = ""
        for line in resp.iter_lines(decode_unicode=True):
            if not line:
                continue
            if line.startswith("event: "):
                event_name = line[7:].strip()
            elif line.startswith("data: "):
                try:
                    payload = json.loads(line[6:])
                except json.JSONDecodeError:
                    continue
                yield event_name, payload

    def _send_message_stream(
        self,
        text: str,
        *,
        session_id: str,
        domain_id: str,
    ) -> MessageResult:
        body = {"session_id": session_id, "domain_id": domain_id, "text": text}
        resp = self._request(
            "POST",
            "/message",
            params={"stream": "1"},
            json_body=body,
            stream=True,
        )
        if resp.status_code >= 400:
            raise GroundedAPIError(f"Stream failed: HTTP {resp.status_code}", status_code=resp.status_code)
        streamed: list[str] = []
        done_payload: dict[str, Any] = {}
        for event_name, payload in self._iter_sse_events(resp):
            if event_name == "token":
                chunk = payload.get("text") or ""
                if chunk:
                    streamed.append(str(chunk))
            elif event_name == "error":
                raise GroundedAPIError(str(payload.get("error") or "stream error"), payload=payload)
            elif event_name == "done":
                done_payload = payload
                break
        messages = list(done_payload.get("messages") or [])
        if not messages:
            messages = self.history(session_id)
        return MessageResult(
            session_id=str(done_payload.get("session_id", session_id)),
            domain_id=str(done_payload.get("domain_id", domain_id)),
            tenant_id=str(done_payload.get("tenant_id", self.tenant_id)),
            messages=messages,
            success=True,
            raw={"streamed_text": "".join(streamed), **done_payload},
        )

    def stream_message_deltas(
        self,
        text: str,
        *,
        session_id: str,
        domain_id: str = "default",
    ) -> Iterator[str]:
        """Yield SSE text deltas from POST /message?stream=1."""
        body = {"session_id": session_id, "domain_id": domain_id, "text": text}
        resp = self._request(
            "POST",
            "/message",
            params={"stream": "1"},
            json_body=body,
            stream=True,
        )
        if resp.status_code >= 400:
            raise GroundedAPIError(f"Stream failed: HTTP {resp.status_code}", status_code=resp.status_code)
        for event_name, payload in self._iter_sse_events(resp):
            if event_name == "token":
                chunk = payload.get("text") or ""
                if chunk:
                    yield str(chunk)
            elif event_name == "error":
                raise GroundedAPIError(str(payload.get("error") or "stream error"), payload=payload)
            elif event_name == "done":
                break

    def feedback(self, message_id: int, rating: int, *, session_id: str | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"message_id": message_id, "rating": rating}
        if session_id:
            body["session_id"] = session_id
        return self._parse_json(self._request("POST", "/feedback", json_body=body))

    def chat(self, text: str, *, domain_id: str = "default", session_id: str | None = None) -> MessageResult:
        """Create session if needed, send one message, return result."""
        sid = session_id or self.create_session(domain_id=domain_id)
        return self.send_message(text, session_id=sid, domain_id=domain_id)
