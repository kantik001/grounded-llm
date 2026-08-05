# `site/` — GitHub Pages landing

Static marketing page for the project (not the product chat UI — that is [`webapp/`](../webapp/README.md)).

| File | Role |
|------|------|
| `index.html` | Landing (release highlights, links) |
| `style.css` | Styles |
| `packs.json` | Pack registry snapshot for the page |

Regenerate pack list:

```bash
python scripts/build_site_data.py
```

Deployed by [`.github/workflows/pages.yml`](../.github/workflows/pages.yml) from this directory.
