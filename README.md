# go-srvc website

Source for [go-srvc.com](https://go-srvc.com).

Static site built with a small Go generator. Plain HTML/CSS/JS, no JS framework, no npm.

## Build

```sh
make build           # generates dist/
make test            # go test ./...
```

Open `dist/index.html` directly in a browser to preview. Pages use relative paths so `file://` works without a server.

## Deploy

Pushes to `main` deploy automatically via the `Deploy` workflow.
GitHub Pages must be configured with **Source: GitHub Actions** in the repo settings.

## Layout

```
cmd/gen/             # site generator (Go)
templates/           # html/template sources
assets/              # css, js, images copied to dist/assets
static/              # files copied to dist/ root (CNAME, robots.txt, etc.)
```
