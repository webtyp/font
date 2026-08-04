# Specification — `tinywasm/font`

Strict functional requirements: exact public surface, exact derivation tables,
exact failure conditions. Structure and reasoning are not repeated here — see
[ARCHITECTURE.md](ARCHITECTURE.md).

Every table below is a test assertion. An implementation is correct when it matches
these tables byte for byte.

---

## 1. Public surface

```go
// Family is the name of the product's typeface.
type Family string

// Style names a face. The set is closed: these four, and exactly the four
// tinywasm/pdf registers.
type Style uint8

const (
    Regular   Style = iota // 0
    Bold                   // 1
    Italic                 // 2
    BoldItalic             // 3
)

// Face is a face's file name, derived — never written by hand. Without
// extension: each medium adds its own (.ttf for PDF, .woff2 for web).
func (f Family) Face(s Style) string

// Declaration is what a project declares in its font.go. It can only be built
// with Declare, so a partial declaration cannot exist.
type Declaration struct{ /* unexported fields */ }

func Declare(family Family, dir string) Declaration
func (d Declaration) Family() Family
func (d Declaration) Dir() string
```

Invariants:

- No exported field on `Declaration`; access only via `Family()` and `Dir()`.
- `Declare` never returns an error and never panics.
- No `map[`, `[]byte`, `os.` or `embed` anywhere in the package's non-test
  source, and no build directive in either form — enforced by
  `TestRootIsWasmSafe`. Test files are exempt: they never enter a binary.
- Importing this root is free and safe from any medium; importing a `font`
  subpackage is a backend decision, and the file that does it carries
  `//go:build !wasm`. The root never imports its own subpackages — that arrow
  only points outward.

---

## 2. Face derivation

With `f = Family("Roboto")`:

| `Style` | `f.Face(s)` |
|---|---|
| `Regular` | `Roboto-Regular` |
| `Bold` | `Roboto-Bold` |
| `Italic` | `Roboto-Italic` |
| `BoldItalic` | `Roboto-BoldItalic` |

Rules:

1. `Face` returns exactly the strings above, **without** extension or directory.
2. A `Style` value outside the four constants yields `Regular`'s result — the
   derived name can never be a suffix-less family of an unrelated face.
3. Derivation is deterministic and allocation-free in shape: every face is the
   family plus a `-Style` suffix — the family alone is never a face name.

For any non-empty family, the four derived names are pairwise distinct.

---

## 3. Declaration

With `d = Declare(Family("Roboto"), "fonts/")`:

| Call | Result |
|---|---|
| `Declare("Roboto", "fonts/").Family()` | `Family("Roboto")` |
| `Declare("Roboto", "fonts/").Dir()` | `"fonts/"` |
| `Declare("", "fonts/").Family()` | `Family("")` — constructible; the runtime gives no error |
| `Declare("Roboto", "").Dir()` | `""` |

`Dir` is returned exactly as declared; this package neither adds nor strips a
trailing slash.

`Dir()` is *where the faces live, in the medium of whoever reads it* — it is
relative, not canonical, and there is no universal root:

| Consumer | Medium | Origin it hangs from |
|---|---|---|
| `assetmin` (build) | disk | the project root (`Config.RootDir`) |
| `pdf` in CLI | disk | the working directory |
| `pdf` in WASM | HTTP | the page origin |

What never varies is `Family()`: it is global — the single origin that keeps web
and PDF looking alike — while `Dir()` is per medium.

---

## 4. Consumer-shaped proof

A test in this library that walks a consumer's real path must, with
`d := font.Declare("Roboto", "fonts/")`, assert exactly the lines below, built the
same way a consumer builds them — plain concatenation of `Dir()` + `Face()` +
medium extension. This mirrors what `tinywasm/pdf` registers and what
`tinywasm/assetmin` serves.

PDF view:

| consumer line | value |
|---|---|
| `d.Dir() + d.Family().Face(Regular) + ".ttf"` | `fonts/Roboto-Regular.ttf` |
| `d.Dir() + d.Family().Face(Bold) + ".ttf"` | `fonts/Roboto-Bold.ttf` |
| `d.Dir() + d.Family().Face(Italic) + ".ttf"` | `fonts/Roboto-Italic.ttf` |
| `d.Dir() + d.Family().Face(BoldItalic) + ".ttf"` | `fonts/Roboto-BoldItalic.ttf` |

Web view — the same pairs, `.ttf` replaced by `.woff2`:

| consumer line | value |
|---|---|
| `d.Dir() + d.Family().Face(Regular) + ".woff2"` | `fonts/Roboto-Regular.woff2` |
| `d.Dir() + d.Family().Face(Bold) + ".woff2"` | `fonts/Roboto-Bold.woff2` |
| `d.Dir() + d.Family().Face(Italic) + ".woff2"` | `fonts/Roboto-Italic.woff2` |
| `d.Dir() + d.Family().Face(BoldItalic) + ".woff2"` | `fonts/Roboto-BoldItalic.woff2` |

---

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure and constraints these values serve.
- [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md) — the ecosystem principles this piece follows.