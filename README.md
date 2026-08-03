# font
<img src="docs/img/badges.svg">

Typeface identity shared by web and PDF: WASM-safe names, no embedded bytes.

A product declares its typeface once; every medium derives the face names from
that single origin.

```go
func (m Module) RenderFonts() font.Declaration {
    return font.Declare("Roboto", "fonts/")
}
```

## Documentation

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — what this module is, why it exists, and where it sits in the suite.
- [docs/SPECS.md](docs/SPECS.md) — exact public surface and face derivation tables.
- [docs/CONSTRUCTION_HARNESS.md](docs/CONSTRUCTION_HARNESS.md) — the ecosystem principles this piece follows.
