"""Tests for HTML to text helper."""

from connectors._html import html_to_text


def test_html_to_text_strips_tags():
    raw = "<p>Hello <strong>world</strong></p><ul><li>One</li></ul>"
    text = html_to_text(raw)
    assert "Hello world" in text
    assert "One" in text
    assert "<" not in text
