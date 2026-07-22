#!/usr/bin/env python3
"""Generate README demo GIF — chat UI mockup with real API data when available."""

from __future__ import annotations

import json
import sys
import textwrap
import urllib.error
import urllib.request
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "assets" / "demo.gif"

W, H = 720, 900
HEADER_BG = (42, 171, 238)
BG = (229, 221, 213)
USER_BUBBLE = (220, 248, 198)
BOT_BUBBLE = (255, 255, 255)
CITE_BG = (240, 248, 255)
CITE_BORDER = (42, 171, 238)
TEXT = (17, 17, 17)
HINT = (112, 111, 111)
WHITE = (255, 255, 255)


def _font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    candidates = [
        "C:/Windows/Fonts/segoeui.ttf",
        "C:/Windows/Fonts/segoeuib.ttf" if bold else "C:/Windows/Fonts/segoeui.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/System/Library/Fonts/Supplemental/Arial.ttf",
    ]
    for path in candidates:
        try:
            return ImageFont.truetype(path, size)
        except OSError:
            continue
    return ImageFont.load_default()


def fetch_chat() -> tuple[str, str, str]:
    """Return (question, answer, citation_filename) from live API or fallback."""
    question = "How many paid vacation days do employees get?"
    answer = "Employees receive 28 paid vacation days per year."
    citation = "vacation_policy_en.txt"
    base = "http://127.0.0.1:8080"
    try:
        req = urllib.request.Request(
            f"{base}/api/session",
            data=json.dumps({"domain_id": "default"}).encode(),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=5) as resp:
            session_id = json.loads(resp.read().decode())["session_id"]
        msg_body = json.dumps(
            {"session_id": session_id, "domain_id": "default", "text": question}
        ).encode()
        req2 = urllib.request.Request(
            f"{base}/api/message",
            data=msg_body,
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(req2, timeout=15) as resp:
            data = json.loads(resp.read().decode())
        for m in reversed(data.get("messages", [])):
            if m.get("role") == "assistant":
                answer = (m.get("content") or answer).split("\n\n")[0].strip()
                cites = m.get("citations") or []
                if cites:
                    citation = cites[0].get("filename") or citation
                break
    except (urllib.error.URLError, TimeoutError, KeyError, json.JSONDecodeError):
        pass
    return question, answer, citation


def _draw_header(draw: ImageDraw.ImageDraw, title_font, sub_font) -> None:
    draw.rectangle([0, 0, W, 130], fill=HEADER_BG)
    draw.text((20, 22), "Grounded LLM", font=title_font, fill=WHITE)
    draw.text((20, 52), "Answers grounded in your knowledge base", font=sub_font, fill=WHITE)
    draw.rounded_rectangle([20, 82, 260, 112], radius=8, fill=WHITE)
    draw.text((32, 90), "Domain: HR Policies", font=sub_font, fill=TEXT)


def _bubble(
    draw: ImageDraw.ImageDraw,
    x: int,
    y: int,
    text: str,
    *,
    user: bool,
    font,
    max_w: int = 420,
) -> int:
    lines: list[str] = []
    for paragraph in text.split("\n"):
        lines.extend(textwrap.wrap(paragraph, width=38) or [""])
    line_h = 22
    pad_x, pad_y = 16, 12
    bubble_w = min(max_w, max((font.getbbox(line)[2] for line in lines), default=120) + pad_x * 2)
    bubble_h = len(lines) * line_h + pad_y * 2
    color = USER_BUBBLE if user else BOT_BUBBLE
    bx = x if user else x
    if user:
        bx = W - bubble_w - 24
    draw.rounded_rectangle([bx, y, bx + bubble_w, y + bubble_h], radius=14, fill=color)
    ty = y + pad_y
    for line in lines:
        draw.text((bx + pad_x, ty), line, font=font, fill=TEXT)
        ty += line_h
    return y + bubble_h + 16


def _citation_box(draw: ImageDraw.ImageDraw, y: int, filename: str, font, small) -> int:
    box_h = 72
    draw.rounded_rectangle([44, y, W - 44, y + box_h], radius=10, outline=CITE_BORDER, width=2, fill=CITE_BG)
    draw.text((58, y + 12), "Source citation", font=small, fill=HINT)
    draw.text((58, y + 32), filename, font=font, fill=CITE_BORDER)
    return y + box_h + 12


def _composer(draw: ImageDraw.ImageDraw, text: str, font) -> None:
    draw.rectangle([0, H - 88, W, H], fill=WHITE)
    draw.rounded_rectangle([16, H - 72, W - 72, H - 24], radius=22, fill=(245, 245, 245))
    draw.text((32, H - 58), text or "Message…", font=font, fill=HINT if not text else TEXT)
    draw.ellipse([W - 60, H - 64, W - 16, H - 20], fill=HEADER_BG)
    draw.text((W - 44, H - 54), ">", font=font, fill=WHITE)


def frame_base(title_font, sub_font) -> tuple[Image.Image, ImageDraw.ImageDraw]:
    img = Image.new("RGB", (W, H), BG)
    draw = ImageDraw.Draw(img)
    _draw_header(draw, title_font, sub_font)
    return img, draw


def build_frames(question: str, answer: str, citation: str) -> list[Image.Image]:
    title = _font(22, bold=True)
    sub = _font(14)
    body = _font(16)
    small = _font(13)
    frames: list[Image.Image] = []

    # Frame 1 — empty chat + sample chip
    img, draw = frame_base(title, sub)
    draw.text((24, 150), "Chat with assistant", font=small, fill=HINT)
    draw.rounded_rectangle([24, 180, 340, 214], radius=16, fill=BOT_BUBBLE)
    draw.text((40, 192), "How many vacation days?", font=small, fill=HINT)
    _composer(draw, "", body)
    frames.append(img)

    # Frame 2 — user question typed
    img, draw = frame_base(title, sub)
    y = 160
    y = _bubble(draw, 24, y, question, user=True, font=body)
    _composer(draw, "", body)
    frames.append(img)

    # Frame 3 — typing indicator
    img, draw = frame_base(title, sub)
    y = 160
    y = _bubble(draw, 24, y, question, user=True, font=body)
    draw.rounded_rectangle([24, y, 160, y + 36], radius=14, fill=BOT_BUBBLE)
    draw.text((40, y + 8), "Assistant is typing…", font=small, fill=HINT)
    _composer(draw, "", body)
    frames.append(img)

    # Frame 4 — answer + citation
    img, draw = frame_base(title, sub)
    y = 160
    y = _bubble(draw, 24, y, question, user=True, font=body)
    y = _bubble(draw, 24, y, answer, user=False, font=body, max_w=480)
    y = _citation_box(draw, y, citation, body, small)
    draw.text((24, y + 4), "Verified: numbers match knowledge base", font=small, fill=(34, 139, 34))
    _composer(draw, "", body)
    frames.append(img)

    # Hold final frame longer via duplication
    frames.extend([img.copy(), img.copy()])
    return frames


def main() -> int:
    question, answer, citation = fetch_chat()
    OUT.parent.mkdir(parents=True, exist_ok=True)
    frames = build_frames(question, answer, citation)
    durations = [900, 900, 700, 1400, 1800, 1800]
    frames[0].save(
        OUT,
        save_all=True,
        append_images=frames[1:],
        duration=durations,
        loop=0,
        optimize=True,
    )
    print(f"Wrote {OUT} ({OUT.stat().st_size // 1024} KB)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
