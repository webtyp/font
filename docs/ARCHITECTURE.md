# Architecture of `webtyp/font`

Defines the **what** and **why** of typeface identity: the single origin from which
web and PDF derive their font names. Abstract structure only — the exact API surface
and the derivation tables live in [SPECS.md](SPECS.md).

---

## 1. What `webtyp/font` is

The module that **names the typeface of a product** and **derives the names of its
faces**. Nothing else.

It exists to make one claim true: *a product's typography is decided once.* Today the
same decision is written twice, as a loose string in both ends — the `font-family`
literal inside the `webtyp/css` reset and the font paths registered with
`webtyp/pdf`. Nothing stops web and PDF of the same document from shipping
different typefaces; today that is what happens. This module removes the second
place where the name is written.

It does **not** read, serve, subset or embed font files. It knows nothing about CSS
or PDF. It ships no typeface of its own: every project declares its own.

---

## 2. Position in the suite

One identity crosses two boundaries, and the crossing is what makes it a piece.

| Module | Owns | Never does |
|---|---|---|
| `webtyp/font` | **Identity** — what the product's typeface is called, and the names of its four faces | Read a file, know a medium, own a byte |
| `webtyp/css` | **Values** — the `--font-sans` token fed by the family name | Derive a face name |
| `webtyp/pdf` | **Delivery** — registers face files for rendering | Invent a face name |
| `webtyp/sitec` | **Delivery** — ships the face files to the browser | Invent a face name |

The direction of dependency is fixed: `css`, `pdf` and `sitec` import `font`;
`font` imports nobody.

### 2.1 The partition WASM demands

Two requirements of the product, apparently opposed:

1. **No typeface is ever embedded.** Every project declares its own; no library
   ships one by default.
2. **The WASM binary only ever loads names, never bytes.** The frontend needs to
   *know* which face to request; it does not need to carry it.

They resolve with the partition the ecosystem already uses:

| | `webtyp/widget` | `webtyp/css` | **`webtyp/font`** |
|---|---|---|---|
| Build tag | none | `//go:build !wasm` | **none** |
| Content | identity | values | **identity** |
| Crosses to WASM | yes | no | **yes** |

`widget` is the exact precedent: not a single build tag, only identity types. This
piece lives on that side of the line, and so does the project's own `config/fonts.go`
— a family name and a folder name are identity, and the frontend needs both to
request a face (§5). What stays on the `!wasm` side are the **values**: the stylesheet
in `config/css.go`, and the font files themselves, which no Go file ever carries.

**Review rule:** if a file of this package ever needs a build tag, the piece is
wrongly split. Pure identity does not need one.

---

## 3. The contract

Four concepts, each closing a failure mode of the loose-string era:

- **`Family`** is the product's typeface name, a plain named string. It is the only
  thing that crosses to WASM.
- **`Style`** is a closed enum of four faces — exactly the four `webtyp/pdf`
  registers. A wrong style cannot be written: `"i"`, `"Italic"` or `"BI"` do not
  compile.
- **`Face`** is a *derived* file name, never written by hand. Deriving is what makes
  web and PDF unable to diverge: there is a single origin of the name. The
  extension is deliberately absent — the extension belongs to the medium (`.woff2`
  for web, `.ttf` for PDF), not to the identity.
- **`Declaration`** is what a project states in its `font.go`: the family and the
  subfolder where its faces live. It has unexported fields and a single
  construction path, `Declare`, so a partial declaration cannot exist.

---

## 4. Constraints

- **Identity only.** If a `[]byte`, an `os.` call or an `embed` appears, the piece
  left its responsibility.
- **No maps.** Four styles are an array or a switch. A `map[Style]string` adds
  dead weight to a TinyGo binary for zero benefit.
- **No runtime errors.** `Declare` cannot fail: an empty string is not a defect the
  runtime should report, and nothing else is left to validate — the types already
  did the work at compile time.
- **No defaults.** This library does not decide what a product looks like. A
  project that does not declare its typeface gets none.
- **Deterministic.** Deriving the same style twice yields the same name; emission
  order is fixed by the enum values.

---

## 5. How a project uses it

A product declares once, in `config/fonts.go`, as a package-level function:

```go
package config

import "webtyp.com/font"

func Fonts() font.Declaration {
    return font.Declare("Roboto", "fonts/")
}
```

**That file carries no build tag, and the omission is the design.** A `//go:build
!wasm` file is not compiled for the browser, and the frontend is precisely where the
PDF is generated: that code has to know the family is `"Roboto"` to request
`fonts/Roboto-Bold.ttf`. Tagging the declaration would put it out of reach of its own
consumer. A declaration is identity, and identity crosses (§2.1).

Its neighbour `config/css.go` **does** carry `//go:build !wasm`, because it returns a
stylesheet full of values. Both live in the same Go package, so `RootCSS()` calls
`Fonts()` directly — no extraction mechanism is involved, and `webtyp/ssr` needs no
change.

From this single declaration, three consumers get what they need without repeating
the decision:

- `webtyp/css` takes the family for `--font-sans`.
- `webtyp/pdf` derives the four face names and appends `.ttf`.
- `webtyp/sitec` delivers the face files to the browser.

The WASM binary receives `"Roboto"` and the derivation rule — never a font byte.

---

## 6. The boundary with `webtyp/sitec`

`sitec` already owns a typed contract for a binary asset declared in a `!wasm`
file: `ImageProcessor` (implemented by `webtyp/image/min`, injected by the
composition root). Fonts are the same case. `sitec` will expose a `FontProcessor`
pattern, calqued on `ImageProcessor`, and serve the faces `Declaration` names.

The contract at the seam is `Declaration` — a type this package owns. A consumer
never re-creates it locally, and a missing method here is a defect in this library,
not an excuse to patch downstream.

---

## Related documents

- [SPECS.md](SPECS.md) — exact public surface and derivation tables.
- [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md) — the ecosystem principles this piece follows.
