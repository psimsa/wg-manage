# Session handover protocol

How to hand this project between agent sessions (or humans) without losing
context. Update the "Current state" section at the start of every session.

## 1. Repositories, branches, PRs

| Item | Value |
|---|---|
| Repo | `psimsa/wg-manage` (module `github.com/ofcoursedude/wg-manage`) |
| Local clone | `/workspace/github__psimsa__wg-manage` |
| Base branch | `main` |
| Revamp branch | `vibe/revive-wg-manage-f50cd7` → PR **#65** (ready for review) |
| Wasm/PWA branch | `vibe/wasm-pwa-f50cd7` → PR **#66** (draft). **Branched off #65's head** — merge #65 first or rebase #66 onto main after |
| Specs | `docs/specs/001..007` (all marked implemented), `docs/wasm-pwa.md`, this file |

## 2. Current state (as of last update: 2026-11-01, head `dcdf8ca` on `vibe/wasm-pwa-f50cd7`)

- Version: `2.0-preview1` (`internal/buildinfo/version.go`); `version`/`v` command exists.
- CI green: `Go` (build+tests) success on `dcdf8ca`; `Web (wasm/PWA)` success on `16b676a` (test-wasm + build; deploy-pages only on `main`).
- Open issues #59/#34/#16 are implemented in PR #65 (its body has `Closes #…`); #45 (wgctrl vulnerabilities) and #15 (dependency dashboard) are NOT addressed — SonarCloud/CodeFactor checks fail and are disregarded by owner decision.
- **One manual repo setting still required**: Settings → Pages → Source: GitHub Actions (cannot be set via API). Until then, `deploy-pages` on main will fail.

## 3. Architecture map

```
main.go, command.go     CLI entry; Command interface; registers cmd/*
cmd/<name>/             one package per subcommand (add, bootstrap, generate, init,
                        remove, format, recreate, pool, psk, tui, viz, version)
models/                 Configuration/Peer structs (+inline Extra maps for
                        user-defined yaml keys), yaml load/save, conf rendering,
                        settings.go (Settings + Validate + defaults),
                        ipam.go (netip-based allocation), api.go (wasm-safe wrappers)
wg/                     key/PSK generation (wgctrl) + api.go wrappers
tui/                    Bubble Tea model (pure Update, unit-tested), generate.go
viz/                    Render(cfg, ascii|dot|mermaid)
wasm/                   js/wasm entry: global wgManage JS API (build tag js&&wasm)
web/                    PWA shell (index.html, app.js, sw.js, manifest)
scripts/build-wasm.sh   builds web/dist/
internal/buildinfo/     Version constant
acceptance_test.go      root: builds binary, runs CLI scenarios vs goldens
testdata/acceptance/    goldens (UPDATE_GOLDEN=1 regenerates; wg keys normalized to <KEY>/[KEY8])
```

Key invariants:
- `models.LoadYaml` validates settings; unknown yaml keys are preserved via `,inline` Extra maps (spec 006) — never remove those.
- A peer is a server iff it has an endpoint (`Peer.IsServer()`).
- IPAM: servers allocate from `subnets.servers`, clients from `subnets.clients` (defaults: `10.0.2.0/26` and the whole `10.0.2.0/24` pool).
- `generate` flag precedence: CLI flag > `settings.generate` > built-in default.
- TUI and CLI share the same models/IPAM code — never add a parallel model.

## 4. Standard session workflow

1. `cd /workspace/github__psimsa__wg-manage && git fetch && git status` — pick the right branch (see table above); never work on `main`.
2. Read `docs/specs/README.md` + this file's "Current state"; check `gh pr list` / `gh run list` for CI state of the branch head.
3. Make the smallest change; add/extend tests (`go test ./...`, wasm: `GOOS=js GOARCH=wasm go test -c -o /tmp/w.wasm ./wasm/` + node runner; goldens: `UPDATE_GOLDEN=1 go test ./...` then review the diff).
4. Verify: `gofmt -l .` empty, `go vet ./...`, `go test ./...`, `go build ./...`, plus `GOOS=js GOARCH=wasm go build ./wasm/` and `GOOS=windows/darwin go build .` when touching shared code.
5. Commit per phase (message style: `feat:`/`fix:`/`chore:`/`ci:`/`docs:` + scope), push, then **wait for the branch's CI run to complete green before starting the next phase** (poll: `gh run list --repo psimsa/wg-manage --branch <branch> --limit 1`).
6. Update the PR body (`gh pr edit <n> --body-file`) with what changed; keep `Closes #…` only on PRs that actually implement the issue.
7. Update this file's "Current state" section before ending the session.

## 5. Gotchas learned the hard way

- `.gitignore` has `*.yaml` — test fixtures under `testdata/` need the negation rules already present; check `git check-ignore -v <file>` when adding yaml fixtures (this caused a CI-only failure once).
- yaml.v2 `,inline` maps round-trip unknown keys; do not switch models to yaml.v3 without re-running `models/preserve_test.go`.
- Wasm test binaries don't call `main()` — `wasm/main_test.go` has its own `TestMain` calling `registerAPI()`; under Node the Go runtime never exits, so the test runner needs a watchdog `process.exit(0)` (see `.github/workflows/wasm.yml`).
- `gh` cannot enable GitHub Pages source via API (mutating calls blocked) — one manual setting required.
- Renovate/Dependabot PRs (#46, #47) are open; yaml.v2→v3 migration is intentionally NOT done (see gotcha above).

## 6. Open work / ideas

- Merge order: #65 first, then rebase #66 onto main.
- Enable Pages source manually; verify deploy-pages on next main push.
- Address #45 (wgctrl vulns) — needs a wgctrl bump or migration; evaluate separately.
- Possible next features: editable settings fields in the TUI (currently view + add-default only), network tab embedding `viz` in the TUI, per-peer PSK editing in TUI, QR PNGs on desktop only.
