# Empezar con `tinywasm/font`

Guía para configurar la tipografía de un proyecto. Pensada para leerse de una vez,
sin conocer el resto del ecosistema.

---

## 1. El problema

Tu producto tiene dos salidas visuales:

- **La página web**, que se ve con CSS.
- **El PDF**, que se genera con `tinywasm/pdf`.

Las dos tienen que verse con la misma tipografía. Si la web usa Roboto y el PDF usa
DroidSans, el mismo presupuesto se ve de dos maneras distintas según dónde lo mires.

El problema no es elegir la fuente. Es que **hoy hay que escribir su nombre en dos
sitios**, y nada obliga a que coincidan. Basta con que alguien cambie uno y se olvide
del otro.

`tinywasm/font` elimina el segundo sitio.

---

## 2. La idea en una frase

> **Declaras la tipografía una vez. La web y el PDF derivan sus nombres de esa única
> declaración.**

"Derivar" quiere decir que nadie vuelve a teclear `"Roboto-Bold"`. Se calcula.

---

## 3. Cómo funciona la derivación

Es lo único que hace esta librería, y conviene verlo antes de nada:

```go
d := font.Declare("Roboto", "fonts/")

d.Family()                    // "Roboto"
d.Dir()                       // "fonts/"

d.Family().Face(font.Regular)     // "Roboto-Regular"
d.Family().Face(font.Bold)        // "Roboto-Bold"
d.Family().Face(font.Italic)      // "Roboto-Italic"
d.Family().Face(font.BoldItalic)  // "Roboto-BoldItalic"
```

Fíjate en que `Face()` **no devuelve extensión**. La extensión la pone quien usa el
nombre.

```
                 font.Declare("Roboto", "fonts/")
                              │
                              ▼
                  fonts/Roboto-Bold.ttf
                     ┌────────┴────────┐
                     ▼                 ▼
              WEB (@font-face)    PDF (tinywasm/pdf)
                     └──── mismo archivo ────┘
```

**Un solo archivo por cara, usado por los dos.** Ahí está la garantía de que no
diverjan: no hay dos archivos que puedan desincronizarse.

### ¿Por qué TTF y no WOFF2?

WOFF2 es el formato habitual en la web porque comprime mejor. Pero el motor de
`tinywasm/pdf` lee las tablas `glyf`/`loca` de TrueType: **no sabe leer WOFF2**. Y el
navegador sí sabe leer TTF (`format("truetype")`, soportado en todas partes).

Así que TTF es el único formato que sirve a los dos. Y sale ganando, medido sobre
Roboto subseteada al alfabeto latino:

| | Bytes que viajan | Peticiones |
|---|---|---|
| **Un TTF compartido** (servido con brotli) | **16.132 B** | **1** |
| Un WOFF2 para web + un TTF para PDF | 27.890 B | 2 |

El WOFF2 pesa 13.536 B, poco menos que el TTF comprimido. Pero el PDF no lo puede
leer, así que habría que descargar además el TTF entero: casi el doble de bytes y el
doble de peticiones.

La clave es la **caché del navegador**: cuando tu código genera el PDF, pide
`fonts/Roboto-Bold.ttf` — el mismo archivo que el navegador ya descargó para pintar la
página. Es un acierto de caché, no una petición nueva.

> El TTF conserva las tablas de *kerning* y ligaduras. El navegador las usa; el motor
> de PDF las ignora sin que eso rompa nada.

---

## 4. Cómo empezar

### Paso 1 — Los archivos de fuente

Van en una subcarpeta de tu proyecto. Necesitas **cuatro archivos**: una cara por
estilo.

```
config/
└── fonts/
    ├── Roboto-Regular.ttf
    ├── Roboto-Bold.ttf
    ├── Roboto-Italic.ttf
    └── Roboto-BoldItalic.ttf
```

Los nombres **no son libres**: tienen que ser exactamente lo que devuelve `Face()`.
Si tu archivo se llama `Roboto-negrita.ttf`, no se encontrará.

### Sobre el tamaño (opcional, pero recomendable)

Una Roboto completa pesa ~400 KB por cara porque trae más de 3.000 glifos: griego,
cirílico, vietnamita. Si tu producto es en español, recortarla al alfabeto latino la
deja en ~23 KB. Son 17 veces menos.

Ese recorte se llama *subsetting* y se hace una vez, cuando eliges la tipografía. No
es obligatorio: si versionas la fuente completa, todo funciona igual — sólo pesa más.

Si la recortas, **anota junto a los archivos de qué versión salieron y qué rango
conservaste**. Sin esos dos datos nadie puede añadir un glifo que falte más adelante:

```
config/fonts/README.md
──────────────────────
Roboto v3.016 (github.com/googlefonts/roboto-3-classic), caras unhinted.
Subset al rango latino de Google Fonts + € (U+20AC):
U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,U+0304,
U+0308,U+0329,U+2000-206F,U+2074,U+20AC,U+2122,U+2191,U+2193,U+2212,
U+2215,U+FEFF,U+FFFD

Regenerar:  pyftsubset <cara>.ttf --unicodes="<rango>" \
              --layout-features=kern,liga,clig --output-file=<cara>.ttf
```

### Paso 2 — `config/fonts.go`

Aquí declaras. **Este archivo NO lleva build tag.**

```go
package config

import "github.com/tinywasm/font"

// Fonts declara la tipografía del producto. Es el único sitio donde se escribe.
func Fonts() font.Declaration {
    return font.Declare("Roboto", "fonts/")
}
```

### Paso 3 — `config/css.go` lee la familia

Este archivo **sí** lleva `//go:build !wasm`, porque contiene valores CSS.

```go
//go:build !wasm

package config

func RootCSS() *css.Stylesheet {
    return css.Theme(
        css.Set(css.ColorSecondary, "#3f88bf"),
        css.Set(css.FontSans, css.FontStack(Fonts().Family())),  // ← sin repetir el nombre
    )
}
```

Los dos archivos están en el **mismo paquete Go**, así que `css.go` llama a `Fonts()`
directamente. No hace falta ningún mecanismo de extracción.

### Paso 4 — El PDF

```go
tf, err := pdf.LoadDeclared(config.Fonts())
if err != nil {
    return err
}
doc := pdf.NewDocument(tf)
```

`LoadDeclared` deriva las cuatro rutas (`fonts/Roboto-Regular.ttf`, `fonts/Roboto-Bold.ttf`…)
a partir de la declaración. Tú no las escribes.

---

## 5. Por qué `config/fonts.go` no lleva `!wasm`

Esta es la parte que más confunde, y tiene una razón exacta.

Un archivo con `//go:build !wasm` **no se compila para el navegador**. El código WASM
no lo ve, no puede llamarlo.

En tu caso **el PDF se genera en el navegador**. Ese código WASM necesita saber que la
familia es `"Roboto"` para pedir `fonts/Roboto-Bold.ttf`. Si `fonts.go` estuviera
marcado `!wasm`, no podría averiguarlo.

La regla del ecosistema (`css/AGENTS.md` §1) lo dice así:

> Identity strings that the frontend genuinely needs belong to a module that is
> **identity-only and WASM-safe**.

| Qué es | Ejemplo | ¿Lleva `!wasm`? |
|---|---|---|
| **Identidad** — un nombre | `"Roboto"`, `"fonts/"` | **No** |
| **Valores** — cómo se ve algo | `#3f88bf`, `1.5rem` | **Sí** |

`fonts.go` sólo tiene identidad. `css.go` tiene valores. Por eso uno lleva tag y el
otro no.

**¿Y no engorda el binario WASM?** Sí, unos pocos bytes: los strings `"Roboto"` y
`"fonts/"`. Pero son exactamente los que el frontend necesita para hacer su trabajo.
Lo que la regla prohíbe no es pesar, es **pesar sin usarse** — meter colores y tamaños
CSS que el navegador nunca va a leer desde Go.

---

## 6. Ciclo de vida

Dónde ocurre cada cosa:

| Momento | Quién | Qué hace |
|---|---|---|
| **Compilación** | `config/css.go` | Lee `Fonts().Family()` → emite `--font-sans: "Roboto", …` |
| **Compilación** | `assetmin` | Lee la declaración → copia los `.ttf` a la carpeta pública y emite el `@font-face` |
| **Carga de la página** | navegador | Descarga `fonts/Roboto-Bold.ttf` desde la URL del `@font-face` |
| **Ejecución (WASM)** | tu código | Llama a `Fonts()` → deriva `fonts/Roboto-Bold.ttf` |
| **Ejecución (WASM)** | `tinywasm/pdf` | Pide ese `.ttf` por `fetch` — **la caché ya lo tiene** — y lo incrusta |

Los dos caminos —CSS y PDF— parten de la misma llamada a `Fonts()` y terminan en el
mismo archivo. Ahí está la garantía.

---

## 7. Qué va en `config/`, y por qué

| Archivo | Build tag | Para qué |
|---|---|---|
| `config/fonts.go` | *ninguno* | La declaración. Lo lee el build **y** el navegador. |
| `config/fonts/*.ttf` | — | Las cuatro caras. Las pinta el navegador **y** las incrusta el PDF. |
| `config/fonts/README.md` | — | Sólo si recortaste la fuente: versión upstream y rango Unicode. Ver paso 1. |
| `config/css.go` | `!wasm` | Ya existe. Lee la familia y pone la cadena de respaldo. |

Ese README no es burocracia: un `.ttf` subseteado es un **derivado**, y sin saber de
dónde salió es un binario huérfano que nadie puede corregir cuando falte un glifo.

> **Por qué no hay un script aquí.** Sería el mismo script en todos los proyectos, y
> el harness lo prohíbe: *«the glue is written once, in the library that owns it — if
> every application would write the same wiring, that wiring belongs to a piece»*.
> Además el ecosistema no usa bash: sus herramientas son CLIs en Go (`gonew`,
> `gotest`, `gopush`). Si recortar fuentes se vuelve repetitivo entre proyectos, la
> respuesta es un `gofont` en `devflow/cmd/`, no un script copiado en cada repo.

---

## 8. Preguntas que suelen aparecer

**¿Puedo usar una fuente variable (una sola con todos los pesos)?**
No. El motor de PDF ignora los ejes de variación: la negrita saldría idéntica a la
regular. Usa cuatro caras estáticas.

**¿Puedo usar un `.otf`?**
No. Los OTF llevan contornos CFF y el motor de PDF lee `glyf`/`loca` de TrueType. Un
`.otf` no se puede cargar.

**¿Y si mi proyecto no genera PDF?**
Declara igual. `css.go` seguirá leyendo la familia, y no pagas nada extra: si el
frontend nunca llama a `Fonts()`, el compilador lo elimina.

**¿Puedo tener dos tipografías (una de títulos y otra de texto)?**
Hoy no. `Declaration` nombra una. Es deliberado: casi ningún producto necesita dos, y
permitirlo abriría la puerta a que web y PDF elijan distinta.

---

## 9. Lo que todavía no existe

Para ser honestos sobre el estado actual:

| Pieza | Estado |
|---|---|
| `font.Declare`, `Family`, `Style`, `Face` | ✅ implementado |
| `pdf.LoadTypeface(4 rutas)` | ✅ implementado |
| **`pdf.LoadDeclared(font.Declaration)`** | ❌ **falta** — hoy hay que escribir las cuatro rutas a mano, que es justo el problema que esto resuelve |
| **`css.FontStack(font.Family)`** | ❌ falta |
| **`css.FontSans` (el token)** | ❌ falta |
| **`assetmin` sirviendo los `.ttf`** | ❌ falta |

Hasta que existan `LoadDeclared` y `FontStack`, el paso 3 y el paso 4 de esta guía no
compilan: el puente entre la declaración y sus dos consumidores está sin construir.
Eso es lo que cubren los planes de `tinywasm/pdf`, `tinywasm/css` y `tinywasm/assetmin`.

---

## Documentos relacionados

- [ARCHITECTURE.md](ARCHITECTURE.md) — el qué y el porqué de la pieza.
- [SPECS.md](SPECS.md) — la superficie pública exacta.
