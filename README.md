# go-srvc website

Source for [go-srvc.com](https://go-srvc.com).

Static site built with a small Go generator. Plain HTML/CSS/JS, no JS framework, no npm.

## Build

```sh
make build           # generates dist/ (clones every tagged release into .cache/)
make test            # go test ./...
make vet             # go vet website + embedded examples
make example-build   # compile the embedded examples
make example-fmt     # gofmt check for the embedded examples
make tidy            # go mod tidy website + embedded examples
```

Open `dist/index.html` directly in a browser to preview. Pages use relative paths so `file://` works without a server.

## Deploy

Pushes to `main` deploy automatically via the `Deploy` workflow.
GitHub Pages must be configured with **Source: GitHub Actions** in the repo settings.

## Layout

```
cmd/gen/                     # site generator entry point
internal/catalog/            # list of documented packages and their git repos
internal/source/             # tag listing and shallow checkouts into .cache/
internal/docparse/           # go/doc extraction (decls, examples, README)
internal/render/             # page rendering, sitemap, syntax CSS
internal/render/templates/   # html/template sources
internal/render/example/     # compiled Go examples shown on the index page
assets/                      # css, js copied to dist/assets
static/                      # files copied to dist/ root (CNAME, robots.txt, llms.txt)
```

Package pages are generated from every semver tag of `go-srvc/srvc` and `go-srvc/mods`; nothing about their APIs is hand-written here.
