# Public launch playbook

Checklist for making **Grounded LLM** public and running a first PR wave.  
Use when the product is ready — repository is currently **private** on GitHub Free (Pages requires public repo or Pro).

---

## Pre-launch (engineering)

- [ ] All CI jobs green on `main` (`eval-retrieval-gate`, `smoke-api`, conformance)
- [ ] Tag release per [RELEASE.md](./RELEASE.md) (e.g. `v0.3.0`)
- [ ] Run `python scripts/init_pack.py registry --validate`
- [ ] Run `python scripts/build_site_data.py` before Pages deploy
- [ ] Review [SECURITY.md](../../SECURITY.md) — no secrets in history (`secret-scan` job)
- [ ] Update README one-liner + demo GIF or screenshot

---

## Go public (GitHub)

1. **Settings → General → Danger zone → Change visibility → Public**
2. Enable Pages: **Settings → Pages → Source: GitHub Actions**
3. Run workflow **Deploy site** manually
4. Verify: `https://<user>.github.io/grounded-llm/`

```powershell
gh repo edit kantik001/grounded-llm --visibility public
gh api -X POST repos/kantik001/grounded-llm/pages -f build_type=workflow
gh workflow run "Deploy site" --ref main
```

---

## PR channels (first wave)

| Channel | Action |
|---------|--------|
| **Hacker News** | Show HN — focus on spec + conformance + on-prem |
| **Reddit** | r/selfhosted, r/MachineLearning, r/LangChain — follow sub rules |
| **LinkedIn / X** | Short demo video, link to spec + benchmark |
| **Dev.to / blog** | Retrieval eval gate article ([blog template](./blog/)) |
| **Product Hunt** | Optional — after stable public site |

Message (one line):

> Open standard for cited, verified document assistants — Spec v1, conformance CLI, on-prem templates.

---

## Post-launch

- [ ] Monitor GitHub Issues / Discussions
- [ ] Partner certification inquiries → [PARTNER_CERTIFICATION.md](./PARTNER_CERTIFICATION.md)
- [ ] Quarterly roadmap review → [GOVERNANCE.md](./GOVERNANCE.md)

---

## Related

- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
- [HIRING.md](../../HIRING.md)
