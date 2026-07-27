# Release process

## Versioning

- **Product tags:** `vMAJOR.MINOR.PATCH` (SemVer) on `main`
- **API path version:** `/api/v1` — see [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md)
- **OpenAPI `info.version`:** `1.0.0` for API v1 (independent from product tag)

Phases **1–11 are merged to `main`**. Latest published tag: **v0.3.0** (local inference, Redis caches, gRPC Retriever, LLM metrics).

---

## Before tagging

1. CI green on `main` (`eval-retrieval-gate`, `smoke-api`, `go-lint`, `python-lint`, `secret-scan`, conformance, `docker-build`)
2. Update [CHANGELOG.md](../../CHANGELOG.md): move `[Unreleased]` → `[VERSION] - YYYY-MM-DD`
3. Run pack registry validation:
   ```bash
   python scripts/init_pack.py registry --validate
   ```
4. Optional benchmark summary:
   ```bash
   python scripts/bench_report.py --version 0.3.0
   ```
5. Build site data before Pages deploy:
   ```bash
   python scripts/build_site_data.py
   ```
6. Confirm docs match the release: [LLM_PROVIDERS.md](./LLM_PROVIDERS.md), [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [COMPATIBILITY.md](./COMPATIBILITY.md), `site/index.html`
7. **GitHub Pages** (repo must be **public** on GitHub Free, or use Pro): Settings → Pages → Source **GitHub Actions**, then run `Deploy site` workflow manually if needed

See [LAUNCH.md](./LAUNCH.md) for public launch checklist.

---

## Tag and release

```bash
git tag -a v0.3.0 -m "v0.3.0 — local LLM providers, Redis caches, gRPC Retriever, token/TTFT metrics"
git push origin v0.3.0
```

The [Release workflow](../../.github/workflows/release.yml) will:

- Publish GHCR images (`grounded-llm-server`, `-python`, `-webapp`)
- Create GitHub Release with notes

---

## Post-release checklist

- [ ] Verify Pages site (`Deploy site` workflow) — landing shows **v0.3.0**
- [ ] Verify GHCR tags `:0.3.0` and `:latest`
- [ ] Run `python -m conformance all --url <prod> --rag-url <rag>` on staging
- [ ] Smoke: `LLM_PROVIDER=ollama` (or cloud) + Redis `X-Cache` + `grpcurl` Retriever if exposed internally
- [ ] Announce: Spec v1 + RFC-0001 + conformance CLI + template packs
- [ ] Optional: dev.to / Show HN per [LAUNCH.md](./LAUNCH.md)

---

## Related

- [PHASE_5.md](./PHASE_5.md) · [PHASE_11.md](./PHASE_11.md)
- [BENCHMARK.md](./BENCHMARK.md)
- [COMPATIBILITY.md](./COMPATIBILITY.md)
- [LLM_PROVIDERS.md](./LLM_PROVIDERS.md)
