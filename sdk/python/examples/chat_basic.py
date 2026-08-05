#!/usr/bin/env python3
"""Minimal SDK example: one cited answer from the HR demo domain."""

from grounded_llm import GroundedClient


def main() -> None:
    client = GroundedClient("http://localhost:8080")
    result = client.chat(
        "How many paid vacation days do employees get?",
        domain_id="default",
    )
    assistant = result.last_assistant_message
    if not assistant:
        print("No assistant reply")
        return
    print(assistant.get("content", ""))
    for cite in assistant.get("citations") or []:
        print(f"  Source: {cite.get('filename', cite)}")


if __name__ == "__main__":
    main()
