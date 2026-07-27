"""One-off: regenerate docs/assets/architecture.png for v0.3."""
from __future__ import annotations

import math
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "assets" / "architecture.png"

W, H = 1600, 1040


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    candidates = [
        r"C:\Windows\Fonts\segoeuib.ttf" if bold else r"C:\Windows\Fonts\segoeui.ttf",
        r"C:\Windows\Fonts\arialbd.ttf" if bold else r"C:\Windows\Fonts\arial.ttf",
    ]
    for p in candidates:
        if Path(p).exists():
            return ImageFont.truetype(p, size)
    return ImageFont.load_default()


def main() -> None:
    img = Image.new("RGB", (W, H), "#ffffff")
    d = ImageDraw.Draw(img)

    f_title = font(34, True)
    f_sec = font(18, True)
    f_box = font(16, True)
    f_sub = font(13)
    f_cap = font(14)
    f_arrow = font(12)

    def rounded(xy, fill, outline, r=14, width=2):
        d.rounded_rectangle(xy, radius=r, fill=fill, outline=outline, width=width)

    def box(xy, fill, outline, title, subtitle=None):
        rounded(xy, fill, outline)
        x0, y0, x1, y1 = xy
        tw = d.textlength(title, font=f_box)
        tx = x0 + (x1 - x0 - tw) / 2
        if subtitle:
            ty = y0 + (y1 - y0) / 2 - 16
            d.text((tx, ty), title, fill="#1a1a2e", font=f_box)
            sw = d.textlength(subtitle, font=f_sub)
            d.text((x0 + (x1 - x0 - sw) / 2, ty + 22), subtitle, fill="#4a5568", font=f_sub)
        else:
            th = 20
            d.text((tx, y0 + (y1 - y0 - th) / 2), title, fill="#1a1a2e", font=f_box)

    def arrow(p1, p2, color="#2d3748", label=None, dashed=False):
        x1, y1 = p1
        x2, y2 = p2
        if dashed:
            length = math.hypot(x2 - x1, y2 - y1) or 1
            dx, dy = (x2 - x1) / length, (y2 - y1) / length
            pos = 0.0
            on = True
            while pos < length - 12:
                seg = 8 if on else 6
                nx = min(pos + seg, length - 12)
                if on:
                    d.line(
                        [(x1 + dx * pos, y1 + dy * pos), (x1 + dx * nx, y1 + dy * nx)],
                        fill=color,
                        width=2,
                    )
                pos = nx
                on = not on
        else:
            d.line([p1, p2], fill=color, width=2)
        ang = math.atan2(y2 - y1, x2 - x1)
        s = 10
        d.polygon(
            [
                (x2, y2),
                (x2 - s * math.cos(ang - 0.4), y2 - s * math.sin(ang - 0.4)),
                (x2 - s * math.cos(ang + 0.4), y2 - s * math.sin(ang + 0.4)),
            ],
            fill=color,
        )
        if label:
            mx, my = (x1 + x2) / 2, (y1 + y2) / 2
            lw = d.textlength(label, font=f_arrow)
            d.rectangle([mx - lw / 2 - 4, my - 10, mx + lw / 2 + 4, my + 8], fill="#ffffff")
            d.text((mx - lw / 2, my - 8), label, fill="#4a5568", font=f_arrow)

    title = "Grounded LLM Architecture (v0.3)"
    tw = d.textlength(title, font=f_title)
    d.text(((W - tw) / 2, 28), title, fill="#1a365d", font=f_title)

    rounded((80, 90, 1520, 210), "#ebf5ff", "#90cdf4", r=18)
    d.text((100, 102), "Clients", fill="#2b6cb0", font=f_sec)
    for x0, y0, x1, y1, t, s in [
        (120, 135, 360, 195, "Web chat", None),
        (390, 135, 680, 195, "Python SDK / REST", None),
        (710, 135, 1000, 195, "Telegram Mini App", None),
        (1030, 135, 1400, 195, "Agents", "gRPC clients"),
    ]:
        box((x0, y0, x1, y1), "#ffffff", "#63b3ed", t, s)

    rounded((80, 260, 1520, 470), "#ebf8ff", "#63b3ed", r=18)
    d.text((100, 272), "Go server :8080", fill="#2b6cb0", font=f_sec)
    for x0, y0, x1, y1, t, s in [
        (120, 320, 320, 410, "Auth", None),
        (360, 320, 600, 410, "REST API v1", None),
        (640, 320, 920, 410, "LLM orchestration", "tokens · TTFT · cache"),
        (960, 320, 1180, 410, "Numeric verify", None),
        (1220, 320, 1440, 410, "Admin", None),
    ]:
        box((x0, y0, x1, y1), "#ffffff", "#4299e1", t, s)
    for a, b in [((320, 365), (360, 365)), ((600, 365), (640, 365)), ((920, 365), (960, 365))]:
        arrow(a, b)
    d.text(
        (100, 430),
        "/metrics  ·  X-Cache: HIT|MISS on cached answers",
        fill="#4a5568",
        font=f_sub,
    )

    arrow((240, 195), (220, 320))
    arrow((535, 195), (480, 320))
    arrow((855, 195), (780, 320))

    rounded((60, 540, 470, 900), "#f0fff4", "#68d391", r=18)
    d.text((80, 555), "Python RAG", fill="#276749", font=f_sec)
    box((85, 595, 445, 665), "#ffffff", "#48bb78", "HTTP :5000", "/rag/context")
    box((85, 685, 445, 755), "#ffffff", "#48bb78", "gRPC Retriever :50051", "grounded.rag.v1")
    box((85, 775, 445, 865), "#ffffff", "#48bb78", "Hybrid BM25 + RRF", "Chroma / Qdrant / pgvector")

    rounded((500, 540, 820, 820), "#faf5ff", "#b794f4", r=18)
    d.text((520, 555), "Storage", fill="#553c9a", font=f_sec)
    box((520, 595, 800, 680), "#ffffff", "#9f7aea", "PostgreSQL", "sessions · messages")
    box((520, 710, 800, 800), "#ffffff", "#9f7aea", "Files", "data/{tenant}/{domain}")

    rounded((850, 540, 1170, 820), "#fffaf0", "#ed8936", r=18)
    d.text((870, 555), "Redis :6379", fill="#c05621", font=f_sec)
    box((870, 595, 1150, 680), "#ffffff", "#dd6b20", "Embedding cache", "Python · TTL 1h")
    box((870, 710, 1150, 800), "#ffffff", "#dd6b20", "Response cache", "Go · X-Cache · TTL 24h")

    rounded((1200, 540, 1540, 860), "#fffff0", "#d69e2e", r=18)
    d.text((1220, 555), "LLM (OpenAI-compatible)", fill="#975a16", font=f_sec)
    box((1220, 595, 1520, 665), "#ffffff", "#ecc94b", "Cloud / OpenRouter", "LLM_PROVIDER=openai")
    box((1220, 685, 1520, 755), "#ffffff", "#ecc94b", "Ollama", "--profile ollama")
    box((1220, 775, 1520, 845), "#ffffff", "#ecc94b", "vLLM", "--profile vllm")

    arrow((480, 410), (265, 595), label="/rag/context")
    # Agents → gRPC Retriever (around the right edge)
    pts = [(1215, 195), (1560, 195), (1560, 720), (445, 720)]
    for i in range(len(pts) - 1):
        d.line([pts[i], pts[i + 1]], fill="#718096", width=2)
    d.polygon([(445, 720), (457, 714), (457, 726)], fill="#718096")
    gw = d.textlength("gRPC", font=f_arrow)
    d.rectangle([1490 - gw / 2 - 4, 440, 1490 + gw / 2 + 4, 460], fill="#ffffff")
    d.text((1490 - gw / 2, 442), "gRPC", fill="#4a5568", font=f_arrow)

    arrow((780, 410), (1010, 710), label="response cache")
    arrow((780, 410), (660, 595))
    arrow((780, 410), (1360, 595), label="/v1/chat/completions")
    arrow((1330, 410), (660, 750))
    # embeddings cache: under the bottom row
    epts = [(265, 865), (265, 920), (1010, 920), (1010, 800)]
    for i in range(len(epts) - 1):
        d.line([epts[i], epts[i + 1]], fill="#2d3748", width=2)
    d.polygon([(1010, 800), (1004, 812), (1016, 812)], fill="#2d3748")
    ew = d.textlength("embeddings", font=f_arrow)
    d.rectangle([620 - ew / 2 - 4, 902, 620 + ew / 2 + 4, 922], fill="#ffffff")
    d.text((620 - ew / 2, 904), "embeddings", fill="#4a5568", font=f_arrow)

    cap = "v0.3 — local inference · Redis caches · gRPC Retriever · hybrid BM25+RRF · Prometheus metrics"
    cw = d.textlength(cap, font=f_cap)
    d.text(((W - cw) / 2, 980), cap, fill="#718096", font=f_cap)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    img.save(OUT, "PNG", optimize=True)
    print(f"wrote {OUT} ({OUT.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
