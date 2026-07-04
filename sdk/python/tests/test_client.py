"""Tests for grounded_llm client."""

import responses

from grounded_llm import GroundedClient
from grounded_llm.exceptions import GroundedAPIError


@responses.activate
def test_create_session_and_message():
    base = "http://localhost:8080"
    responses.add(responses.GET, f"{base}/health", json={"status": "healthy"}, status=200)
    responses.add(
        responses.POST,
        f"{base}/api/v1/session",
        json={"success": True, "session_id": "sess-1", "domain_id": "default"},
        status=200,
    )
    responses.add(
        responses.POST,
        f"{base}/api/v1/message",
        json={
            "success": True,
            "session_id": "sess-1",
            "domain_id": "default",
            "tenant_id": "default",
            "messages": [
                {"role": "user", "content": "How many vacation days?"},
                {
                    "role": "assistant",
                    "content": "28 paid vacation days per year.",
                    "citations": [{"filename": "vacation_policy_en.txt"}],
                },
            ],
        },
        status=200,
    )

    client = GroundedClient(base)
    sid = client.create_session("default")
    assert sid == "sess-1"
    result = client.send_message("How many vacation days?", session_id=sid)
    assert result.success
    assert result.last_assistant_message["content"].startswith("28")
    assert result.last_assistant_message["citations"][0]["filename"] == "vacation_policy_en.txt"


@responses.activate
def test_chat_one_shot():
    base = "http://localhost:8080"
    responses.add(
        responses.POST,
        f"{base}/api/v1/session",
        json={"success": True, "session_id": "s2", "domain_id": "default"},
        status=200,
    )
    responses.add(
        responses.POST,
        f"{base}/api/v1/message",
        json={
            "success": True,
            "session_id": "s2",
            "domain_id": "default",
            "tenant_id": "default",
            "messages": [{"role": "assistant", "content": "Answer"}],
        },
        status=200,
    )
    client = GroundedClient(base)
    result = client.chat("Hello")
    assert result.session_id == "s2"


@responses.activate
def test_api_error_raises():
    base = "http://localhost:8080"
    responses.add(
        responses.POST,
        f"{base}/api/v1/session",
        json={"success": False, "error": "Unauthorized"},
        status=401,
    )
    client = GroundedClient(base)
    try:
        client.create_session()
        assert False, "expected error"
    except GroundedAPIError:
        pass


@responses.activate
def test_list_domains():
    base = "http://localhost:8080"
    body = {"success": True, "default_domain": "default", "domains": []}
    responses.add(responses.GET, f"{base}/api/domains", json=body, status=200)
    client = GroundedClient(base)
    data = client.list_domains()
    assert data["default_domain"] == "default"
