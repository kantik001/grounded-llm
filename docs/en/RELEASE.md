# Release process

## Versioning

- **Product tags:** `vMAJOR.MINOR.PATCH` (SemVer) on `main`
- **API path version:** `/api/v1` — see [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md)
- **OpenAPI `info.version`:** `1.0.0` for API v1 (independent from product tag)

## Before tagging `v0.3.0`

1. Merge `feature/phase-5-standard-publication` to `main`
2. CI green on `main` (all jobs including `secret-scan`, `eval-retrieval-gate`)
3. Update [CHANGELOG.md](../../CHANGELOG.md): move `[Unreleased]` → `[0.3.0] - YYYY-MM-DD`
4. Optional: run benchmark locally and attach summary:
   ```bash
   python scripts/bench_report.py --version 0.3.0
   ```
5. Enable **GitHub Pages** (Settings → Pages → Source: **GitHub Actions**)

## Tag and release

```bash
git tag -a v0.3.0 -m "v0.3.0 — standard publication (spec v1, conformance, adversarial eval)"
git push origin v0.3.0
```

The [Release workflow](../../.github/workflows/release.yml) will:

- Publish GHCR images (`grounded-llm-server`, `-python`, `-webapp`)
- Create GitHub Release with notes

## Post-release checklist

- [ ] Verify Pages site deployed (`Deploy site` workflow)
- [ ] Run `python -m conformance all --url <prod> --rag-url <rag>` on staging
- [ ] Announce: Spec v1 + RFC-0001 + conformance CLI

## Related

- [PHASE_5.md](./PHASE_5.md)
- [BENCHMARK.md](./BENCHMARK.md)
- [COMPATIBILITY.md](./COMPATIBILITY.md)
