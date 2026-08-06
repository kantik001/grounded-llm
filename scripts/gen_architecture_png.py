"""Fallback generator for docs/assets/architecture.png.

The canonical README diagram is the icon-rich PNG in docs/assets/ (designed
for readability). Run this script only to regenerate a text-only fallback:

    python scripts/gen_architecture_png.py
"""
from __future__ import annotations

import math
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "assets" / "architecture.png"

W, H = 1536, 1024
BG = "#eef2f7"
BLUE = "#2563eb"
TEAL = "#0d9488"
SLATE = "#1e293b"
MUTED = "#64748b"
WHITE = "#ffffff"


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    candidates = [
        r"C:\Windows\Fonts\segoeuib.ttf" if bold else r"C:\Windows\Fonts\segoeui.ttf",
        r"C:\Windows\Fonts\arialbd.ttf" if bold else r"C:\Windows\Fonts\arial.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
    ]
    for p in candidates:
        if Path(p).exists():
            return ImageFont.truetype(p, size)
    return ImageFont.load_default()


def rounded(d: ImageDraw.ImageDraw, xy, fill, outline, r=12, width=2):
    d.rounded_rectangle(xy, radius=r, fill=fill, outline=outline, width=width)


def arrow(
    d: ImageDraw.ImageDraw,
    p1,
    p2,
    color=BLUE,
    width=2,
    label: str | None = None,
    label_font=None,
    dashed=False,
):
    x1, y1 = p1
    x2, y2 = p2
    if dashed:
        length = math.hypot(x2 - x1, y2 - y1) or 1
        dx, dy = (x2 - x1) / length, (y2 - y1) / length
        pos = 0.0
        on = True
        while pos < length - 14:
            seg = 10 if on else 7
            nx = min(pos + seg, length - 14)
            if on:
                d.line([(x1 + dx * pos, y1 + dy * pos), (x1 + dx * nx, y1 + dy * nx)], fill=color, width=width)
            pos = nx
            on = not on
    else:
        d.line([p1, p2], fill=color, width=width)
    ang = math.atan2(y2 - y1, x2 - x1)
    s = 11
    d.polygon(
        [
            (x2, y2),
            (x2 - s * math.cos(ang - 0.35), y2 - s * math.sin(ang - 0.35)),
            (x2 - s * math.cos(ang + 0.35), y2 - s * math.sin(ang + 0.35)),
        ],
        fill=color,
    )
    if label and label_font:
        mx, my = (x1 + x2) / 2, (y1 + y2) / 2
        lw = d.textlength(label, font=label_font)
        d.rounded_rectangle([mx - lw / 2 - 6, my - 11, mx + lw / 2 + 6, my + 9], radius=4, fill=WHITE)
        d.text((mx - lw / 2, my - 9), label, fill=MUTED, font=label_font)


def label_box(
    d: ImageDraw.ImageDraw,
    xy,
    title: str,
    lines: list[str],
    *,
    outline=BLUE,
    title_font,
    sub_font,
    icon_fn=None,
):
    x0, y0, x1, y1 = xy
    rounded(d, xy, WHITE, outline, r=14)
    cx = (x0 + x1) // 2
    if icon_fn:
        icon_fn(d, cx, y0 + 28)
    ty = y0 + 52 if icon_fn else y0 + 20
    tw = d.textlength(title, font=title_font)
    d.text((cx - tw / 2, ty), title, fill=SLATE, font=title_font)
    ly = ty + 26
    for line in lines:
        lw = d.textlength(line, font=sub_font)
        d.text((cx - lw / 2, ly), line, fill=MUTED, font=sub_font)
        ly += 18


# --- simple line icons (center x, top y) ---


def icon_globe(d, cx, cy):
    d.ellipse([cx - 14, cy - 14, cx + 14, cy + 14], outline=BLUE, width=2)
    d.line([(cx, cy - 14), (cx, cy + 14)], fill=BLUE, width=2)
    d.ellipse([cx - 8, cy - 14, cx + 8, cy + 14], outline=BLUE, width=1)


def icon_code(d, cx, cy):
    d.line([(cx - 12, cy), (cx - 4, cy - 8)], fill=BLUE, width=2)
    d.line([(cx - 12, cy), (cx - 4, cy + 8)], fill=BLUE, width=2)
    d.line([(cx + 12, cy), (cx + 4, cy - 8)], fill=BLUE, width=2)
    d.line([(cx + 12, cy), (cx + 4, cy + 8)], fill=BLUE, width=2)
    d.line([(cx - 3, cy + 10), (cx + 3, cy - 10)], fill=BLUE, width=2)


def icon_plane(d, cx, cy):
    d.polygon([(cx - 12, cy + 6), (cx + 14, cy), (cx - 12, cy - 6)], fill=BLUE)


def icon_robot(d, cx, cy):
    rounded(d, [cx - 12, cy - 10, cx + 12, cy + 10], WHITE, BLUE, r=4)
    d.line([(cx, cy - 10), (cx, cy - 16)], fill=BLUE, width=2)
    d.ellipse([cx - 3, cy - 18, cx + 3, cy - 12], fill=BLUE)
    d.ellipse([cx - 6, cy - 4, cx - 2, cy], fill=BLUE)
    d.ellipse([cx + 2, cy - 4, cx + 6, cy], fill=BLUE)


def icon_shield(d, cx, cy, check=False):
    pts = [(cx, cy - 14), (cx + 12, cy - 6), (cx + 10, cy + 10), (cx, cy + 16), (cx - 10, cy + 10), (cx - 12, cy - 6)]
    d.polygon(pts, outline=BLUE, fill="#eff6ff")
    if check:
        d.line([(cx - 5, cy + 2), (cx - 1, cy + 6), (cx + 7, cy - 4)], fill=TEAL, width=2)


def icon_nodes(d, cx, cy):
    for ox, oy in [(-10, -6), (10, -6), (0, 8)]:
        d.ellipse([cx + ox - 5, cy + oy - 5, cx + ox + 5, cy + oy + 5], fill=BLUE)
    d.line([(cx - 5, cy - 4), (cx + 5, cy - 4)], fill=BLUE, width=1)
    d.line([(cx, cy + 3), (cx - 8, cy - 2)], fill=BLUE, width=1)
    d.line([(cx, cy + 3), (cx + 8, cy - 2)], fill=BLUE, width=1)


def icon_brain(d, cx, cy):
    d.arc([cx - 14, cy - 12, cx, cy + 12], 90, 270, fill=BLUE, width=2)
    d.arc([cx, cy - 12, cx + 14, cy + 12], 270, 90, fill=BLUE, width=2)
    d.line([(cx, cy - 12), (cx, cy + 12)], fill=BLUE, width=1)


def icon_search(d, cx, cy):
    d.ellipse([cx - 12, cy - 12, cx + 4, cy + 4], outline=TEAL, width=2)
    d.line([(cx + 2, cy + 2), (cx + 12, cy + 12)], fill=TEAL, width=2)


def icon_db(d, cx, cy):
    d.ellipse([cx - 14, cy - 12, cx + 14, cy - 2], outline="#7c3aed", width=2)
    d.rectangle([cx - 14, cy - 7, cx + 14, cy + 8], outline="#7c3aed", width=2)
    d.ellipse([cx - 14, cy + 2, cx + 14, cy + 12], outline="#7c3aed", width=2)


def icon_stack(d, cx, cy):
    for i, dy in enumerate([8, 0, -8]):
        rounded(d, [cx - 16, cy + dy - 4, cx + 16, cy + dy + 4], "#fff7ed", "#ea580c", r=3, width=1)


def icon_cloud(d, cx, cy):
    d.ellipse([cx - 16, cy - 2, cx - 2, cy + 10], fill="#eab308")
    d.ellipse([cx - 8, cy - 10, cx + 10, cy + 6], fill="#eab308")
    d.ellipse([cx + 2, cy - 4, cx + 16, cy + 10], fill="#eab308")


def icon_guard(d, cx, cy):
    icon_shield(d, cx, cy, check=False)
    d.line([(cx - 4, cy + 2), (cx, cy + 6), (cx + 6, cy - 4)], fill=TEAL, width=2)


def flow_box(d, xy, title, subtitle, icon_fn, f_title, f_sub, *, icon_kwargs=None):
    x0, y0, x1, y1 = xy
    rounded(d, xy, WHITE, BLUE, r=12)
    cx = (x0 + x1) // 2
    if icon_fn:
        if icon_kwargs:
            icon_fn(d, cx, y0 + 22, **icon_kwargs)
        else:
            icon_fn(d, cx, y0 + 22)
    tw = d.textlength(title, font=f_title)
    d.text((cx - tw / 2, y0 + 48), title, fill=SLATE, font=f_title)
    if subtitle:
        sw = d.textlength(subtitle, font=f_sub)
        d.text((cx - sw / 2, y0 + 72), subtitle, fill=MUTED, font=f_sub)


def main() -> None:
    img = Image.new("RGB", (W, H), BG)
    d = ImageDraw.Draw(img)

    f_title = font(32, True)
    f_band = font(17, True)
    f_box = font(15, True)
    f_sub = font(12)
    f_cap = font(13)
    f_lbl = font(11)

    title = "Grounded LLM Architecture"
    tw = d.textlength(title, font=f_title)
    d.text(((W - tw) / 2, 24), title, fill="#0f172a", font=f_title)

    # --- Clients ---
    d.text((72, 108), "Clients", fill=BLUE, font=f_band)
    clients = [
        (100, 130, 300, 210, "Web chat", icon_globe),
        (330, 130, 560, 210, "Python SDK / REST", icon_code),
        (590, 130, 820, 210, "Telegram Mini App", icon_plane),
        (850, 130, 1100, 210, "Agents (gRPC)", icon_robot),
    ]
    for x0, y0, x1, y1, t, ic in clients:
        flow_box(d, (x0, y0, x1, y1), t, None, ic, f_box, f_sub)

    # --- Go server ---
    rounded(d, (72, 250, 1464, 430), WHITE, "#93c5fd", r=16, width=2)
    d.text((92, 262), "Go server :8080", fill=BLUE, font=f_band)
    go_boxes = [
        (100, 300, 320, 400, "Auth & Admin", icon_shield, {}),
        (360, 300, 580, 400, "REST API v1", icon_nodes, None),
        (620, 300, 900, 400, "LLM orchestration", icon_brain, None),
        (940, 300, 1180, 400, "Numeric verify", icon_shield, {"check": True}),
    ]
    for x0, y0, x1, y1, t, ic, ikw in go_boxes:
        flow_box(d, (x0, y0, x1, y1), t, "local default" if t == "Numeric verify" else None, ic, f_box, f_sub, icon_kwargs=ikw)
    for a, b in [((320, 350), (360, 350)), ((580, 350), (620, 350)), ((900, 350), (940, 350))]:
        arrow(d, a, b, width=2)

    # Optional guardrails beside verify
    label_box(
        d,
        (1220, 300, 1450, 400),
        "grounded-guardrails :50052",
        ["VerifyText (optional)", "numeric + PII rules"],
        outline=TEAL,
        title_font=f_box,
        sub_font=f_sub,
        icon_fn=icon_guard,
    )
    arrow(d, (1180, 350), (1220, 350), color=TEAL, dashed=True, label="optional remote|hybrid", label_font=f_lbl)

    # --- Backends ---
    d.text((72, 468), "Backends & Services", fill=BLUE, font=f_band)

    label_box(
        d,
        (72, 500, 360, 680),
        "Python RAG",
        ["HTTP :5000 · hybrid BM25+RRF", "gRPC Retriever :50051", "Chroma / Qdrant / pgvector"],
        outline=TEAL,
        title_font=f_box,
        sub_font=f_sub,
        icon_fn=icon_search,
    )
    label_box(
        d,
        (390, 500, 620, 680),
        "Storage",
        ["PostgreSQL", "Files · data/{tenant}/{domain}"],
        outline="#7c3aed",
        title_font=f_box,
        sub_font=f_sub,
        icon_fn=icon_db,
    )
    label_box(
        d,
        (650, 500, 880, 680),
        "Redis :6379",
        ["Embedding cache", "LLM response cache (X-Cache)"],
        outline="#ea580c",
        title_font=f_box,
        sub_font=f_sub,
        icon_fn=icon_stack,
    )
    label_box(
        d,
        (910, 500, 1180, 680),
        "LLM (OpenAI-compatible)",
        ["Cloud / OpenRouter", "Ollama profile · vLLM profile"],
        outline="#ca8a04",
        title_font=f_box,
        sub_font=f_sub,
        icon_fn=icon_cloud,
    )
    label_box(
        d,
        (1210, 500, 1464, 680),
        "grounded-guardrails :50052",
        ["VerifyText (optional)", "sibling compose"],
        outline=TEAL,
        title_font=f_box,
        sub_font=f_sub,
        icon_fn=icon_guard,
    )

    # Client → Go (first three only)
    arrow(d, (200, 210), (210, 300))
    arrow(d, (445, 210), (470, 300))
    arrow(d, (705, 210), (760, 300))

    # REST → Python RAG
    arrow(d, (470, 400), (220, 500), label="/rag/context", label_font=f_lbl)

    # Agents → gRPC Retriever (bypass Go)
    path = [(975, 210), (1140, 210), (1140, 640), (360, 640), (360, 680)]
    for i in range(len(path) - 1):
        if i < len(path) - 2:
            d.line([path[i], path[i + 1]], fill=TEAL, width=2)
        else:
            arrow(d, path[i], path[i + 1], color=TEAL, width=2)
    d.rounded_rectangle([1148, 390, 1248, 408], radius=4, fill=WHITE)
    d.text((1155, 392), "gRPC :50051", fill=TEAL, font=f_lbl)

    # LLM orchestration → backends (subtle)
    arrow(d, (760, 400), (760, 460), color=MUTED, width=1)
    arrow(d, (760, 460), (1045, 460), color=MUTED, width=1)
    arrow(d, (1045, 460), (1045, 500), color=MUTED, width=1)
    d.text((820, 442), "/v1/chat/completions", fill=MUTED, font=f_lbl)

    cap = (
        "LOCAL INFERENCE · REDIS CACHES · gRPC RETRIEVER :50051 · "
        "OPTIONAL GUARDRAILS :50052 · HYBRID BM25+RRF"
    )
    cw = d.textlength(cap, font=f_cap)
    d.text(((W - cw) / 2, 960), cap, fill=MUTED, font=f_cap)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    img.save(OUT, "PNG", optimize=True)
    print(f"wrote {OUT} ({OUT.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
