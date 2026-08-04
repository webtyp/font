# tinywasm/font
<img src="docs/img/badges.svg">

Typeface identity shared by web and PDF: WASM-safe names, no embedded bytes.

A product declares its typeface once; every medium derives the face names from
that single origin.

```go
// config/fonts.go — no build tag: identity crosses to WASM
func Fonts() font.Declaration {
    return font.Declare("Roboto", "fonts/")
}
```

**Bring your own faces.** Download them, buy them, or reuse ones you already have —
this module says nothing about where they come from. It requires exactly two things:

- the four files live in the folder you pass to `Declare(...)` — you choose it, and
  `Dir()` is relative to the project root;
- each is named what `Face(Style)` returns, plus `.ttf` — `Roboto-Regular.ttf`,
  `Roboto-Bold.ttf`, `Roboto-Italic.ttf`, `Roboto-BoldItalic.ttf`.

If you have no family chosen yet, [faces/](faces/) offers two verified ones — Roboto
and Inter, subsetted to Latin with `€`, four real faces each. They are files, not
code: no `.go` here embeds them, so they never enter a binary.

Importing this root is free and safe from any medium; importing a `font` subpackage
is a backend decision — the file that does it carries `//go:build !wasm`.

## Documentation

- [docs/EMPEZAR.md](docs/EMPEZAR.md) — guía de inicio en español: ciclo de vida, casos de uso y qué va en `config/`.
- [faces/README.md](faces/README.md) — the bundled typefaces: what they are, where they come from, and how they were subsetted.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — what this module is, why it exists, and where it sits in the suite.
- [docs/SPECS.md](docs/SPECS.md) — exact public surface and face derivation tables.
- [docs/CONSTRUCTION_HARNESS.md](docs/CONSTRUCTION_HARNESS.md) — the ecosystem principles this piece follows.
