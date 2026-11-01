# WebAssembly / PWA build

wg-manage can run entirely in the browser as a Progressive Web App. The Go
core compiles to `js/wasm`; a thin JS API (`wgManage.*`) plus a PWA shell
(HTML + manifest + service worker) wrap it.

## Layout

```
wasm/main.go      js/wasm entry point: registers the wgManage JS API
wasm/api_test.go  API tests, run under Node with GOOS=js GOARCH=wasm
models/api.go     browser-friendly, error-returning wrappers (ParseYaml, ToYaml, RenderPeerConf, RenderAll)
wg/api.go         error-returning key generation wrappers
web/              PWA shell (index.html, app.js, manifest.webmanifest, sw.js)
scripts/build-wasm.sh   builds web/dist/ (wg-manage.wasm + wasm_exec.js + shell)
```

## JS API (global `wgManage`)

Each function takes plain JS values and returns `{ ok: true, result }` or
`{ ok: false, error }`:

- `parseYaml(yaml)` → normalized YAML string (validates + settings check)
- `peerNames(yaml)` → `[{ name, address, server, endpoint }]`
- `generateConf(yaml, peerName)` → `.conf` text for one peer
- `generateAll(yaml)` → `{ peerName: confText }`
- `addPeer(yaml, name, role, endpoint?)` → updated YAML; address auto-allocated from the pool; `role` is `client` or `server` (server requires endpoint)
- `removePeer(yaml, name)` → updated YAML
- `bootstrap(endpoint)` → fresh YAML (server + 2 clients, NAT rules, settings)
- `viz(yaml, format?)` → ASCII / DOT / Mermaid rendering
- `version()` → version string (currently `wg-manage 2.0-preview1`)

All state lives in the YAML text (persisted by the app in `localStorage`), so
the wasm module itself is stateless and there is no filesystem access.

## Build & run locally

```sh
scripts/build-wasm.sh        # produces web/dist/
cd web/dist && python3 -m http.server 8080
# open http://localhost:8080  (PWA installable; service worker caches for offline)
```

> Serving over plain HTTP on localhost is fine for development; for a real
> deployment use HTTPS — service workers and installability require a secure
> context.

## Tests

```sh
GOOS=js GOARCH=wasm go test -c -o /tmp/wasm-tests.wasm ./wasm/
# run with node + wasm_exec.js (see .github/workflows/wasm.yml)
```

The `wasm` package has its own `TestMain` that registers the API before
running tests (test binaries don't call `main()`).

## CI

`.github/workflows/wasm.yml` runs the wasm API tests under Node and builds +
uploads the `wg-manage-web` artifact (`web/dist/`) on pushes touching the
relevant paths. On pushes to `main`, a `deploy-pages` job additionally
publishes `web/dist/` to **GitHub Pages** (source: GitHub Actions) via the
official `configure-pages` / `upload-pages-artifact` / `deploy-pages`
actions. Enable once in repo settings: *Settings → Pages → Build and
deployment → Source: GitHub Actions*. The site then lives at
`https://<owner>.github.io/<repo>/`.

## Limitations / notes

- The wasm binary is ~5 MB uncompressed (Go runtime); gzip gets it to ~1.5 MB.
- QR-code PNG generation is desktop-only (uses `image/png`); the web UI shows
  the `.conf` text instead (paste into any WireGuard client).
- No `wg-quick` execution, obviously — the PWA is for *managing* configs, not
  bringing interfaces up.
