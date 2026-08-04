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

Ready-to-use faces live in [faces/](faces/) — Roboto and Inter, subsetted to Latin
with `€`, four real faces each. Copy the ones you want into your project's
`config/fonts/`. They are files, not code: no `.go` here embeds them, so they never
enter a binary.

## Documentation

- [docs/EMPEZAR.md](docs/EMPEZAR.md) — guía de inicio en español: ciclo de vida, casos de uso y qué va en `config/`.
- [faces/README.md](faces/README.md) — the bundled typefaces: what they are, where they come from, and how they were subsetted.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — what this module is, why it exists, and where it sits in the suite.
- [docs/SPECS.md](docs/SPECS.md) — exact public surface and face derivation tables.
- [docs/CONSTRUCTION_HARNESS.md](docs/CONSTRUCTION_HARNESS.md) — the ecosystem principles this piece follows.
