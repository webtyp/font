# PLAN — `cmd/gofont`: recortar tipografías sin copiar un script por proyecto

## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.** Los principios que gobiernan este trabajo:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 4 | One way to do each thing | Una sola forma de obtener las caras de un proyecto. Hoy hay cero, y la alternativa es que cada repo invente la suya. |
| 5 | Minimal surface | El paquete `font` no cambia. La herramienta es otro paquete del mismo módulo. |
| 6 | Fail at compile time | Lo que no se pueda validar en compilación se valida al recortar, en voz alta: una cara que sale sin `€` es un error, no un aviso. |
| 9 | Lego pieces, never forks | **El motivo del plan.** Sin herramienta, cada proyecto copia el mismo script. |

---

## 1. El problema

Las caras de `faces/` son derivados: se obtienen recortando la fuente original de
~400 KB a ~22 KB. Ese recorte hoy no está automatizado, y su ausencia tiene una forma
concreta y conocida.

Durante el trabajo que originó este módulo se escribió un `regenerar.sh` en
`veltylabs/cotizaciones/print/fonts/`. Funciona. Y es exactamente lo que el harness
prohíbe:

> **The glue is written once, in the library that owns it.** If every application would
> write the same wiring, that wiring belongs to a piece — not to the applications.

Todos los proyectos que usen `webtyp/font` copiarían ese mismo script. Además
arrastra una cadena ajena al ecosistema —Python, `fonttools`, `apt`— que ninguna otra
pieza necesita, y el ecosistema no usa bash: sus nueve herramientas (`gonew`, `gotest`,
`gopush`, …) son CLIs en Go.

**El disparador ya se cumplió:** el mismo recorte se ha hecho a mano para `cotizaciones`
y para las caras de `faces/`.

---

## 2. Por qué la herramienta vive aquí y no en `devflow`

`devflow/cmd/` aloja herramientas de *flujo de trabajo* —crear un repo, correr tests,
publicar—. Recortar una tipografía no es flujo de trabajo: es el concern de este
módulo. Ponerla en `devflow` partiría el tema en dos repos y obligaría a `devflow` a
saber de subsets, rangos Unicode y tablas OpenType.

Principio 9: una responsabilidad, una pieza. **Todo lo tipográfico vive en
`webtyp/font`.**

### Esto no rompe la regla de "identidad pura"

`ARCHITECTURE.md` §4 dice que si aparece un `[]byte`, un `os.` o un `embed`, la pieza
se salió de su responsabilidad. `cmd/gofont` usa las tres cosas. No hay contradicción:

**La regla es sobre el paquete `font`, no sobre el módulo.** `cmd/gofont` es
`package main` — nadie lo importa, nunca entra en un binario ajeno, y no puede
contaminar lo que `font` exporta. Es el mismo reparto que ya usa `webtyp/pdf`, cuyo
`cmd/` no afecta a lo que el frontend enlaza.

**Verificación:** tras el cambio, `grep -rn "\[\]byte\|os\.\|embed" *.go` en la **raíz**
del módulo debe seguir sin encontrar nada.

---

## 3. Qué hace

```
gofont Roboto                    # descarga, recorta e instala en ./config/fonts/
gofont Roboto -o faces/          # destino explícito
gofont Roboto -range=latin+euro  # rango con nombre, no una lista de codepoints
```

Tres pasos, ninguno novedoso: descargar la release upstream, recortar al rango, escribir
las cuatro caras con los nombres que `Family.Face(Style)` deriva.

### Decisiones de diseño

- **Las familias conocidas son un catálogo cerrado en Go**, no un argumento libre.
  `gofont Robotoo` no debe descargar nada ni fallar con un 404: debe decir que esa
  familia no está y listar las que sí. Un `string` libre aquí es el mismo agujero que
  este módulo eliminó en `pdf` (principio 1).
- **El rango se nombra, no se escribe.** `latin+euro` en vez de 20 rangos de
  codepoints. Pegar la lista a mano es la forma de equivocarse en silencio, y el valor
  por defecto debe cubrir el español.
- **Verifica lo que produce.** Tras recortar, comprueba que las cuatro caras existen,
  que cubren `áéíóúñü ¿¡ “” – —` y `€`, y que la cursiva **no** es idéntica a la recta.
  Un subset al que le falta un glifo es un fallo de la herramienta, no un detalle que
  el usuario descubra meses después en un PDF. Es el escalón *loud diagnostic* del
  principio 6.
- **Escribe el `README.md` de destino.** Versión upstream, caras usadas y rango — los
  tres datos sin los cuales las caras quedan huérfanas. Si la herramienta no lo escribe,
  alguien tendrá que acordarse, y eso es un agujero del harness.

### El problema real: no hay subsetter en Go

`golang.org/x/image/font/sfnt` **lee** TrueType pero no recorta. Las opciones, y hay que
elegir explícitamente:

| Opción | Coste | Riesgo |
|---|---|---|
| **A** · Envolver `pyftsubset` | bajo — la herramienta detecta si falta y lo dice | dependencia de Python en la máquina de desarrollo |
| **B** · Subsetter propio en Go | alto — reescribir `glyf`/`loca`/`cmap`/`hmtx` y los `GSUB`/`GPOS` supervivientes | correcto pero caro; es un proyecto en sí |
| **C** · Publicar sólo caras pre-recortadas en `faces/` | ninguno | no resuelve al proyecto que quiere otra familia u otro rango |

**Recomendación: A**, con el fallo explícito. `fonttools` es la implementación de
referencia del formato; reimplementarla para ahorrar una dependencia de desarrollo es
mal negocio. Y si la herramienta detecta su ausencia y dice `sudo apt install
fonttools`, el fallo es de diez segundos, no un misterio.

B queda como puerta abierta: si la dependencia estorba de verdad, se sustituye por
dentro sin tocar la interfaz.

---

## 4. Lo que la herramienta NO hace

- **No convierte a WOFF2.** Sobra: una sola cara TTF sirve a web y PDF, y con brotli
  transfiere 16.132 B contra los 27.890 B del esquema de dos formatos. Está medido en
  `app-releases/docs/TYPOGRAPHY_MASTER_PLAN.md` §5.2.
- **No elige la tipografía del producto.** Eso es `config/fonts.go`.
- **No corre en el build.** Se invoca a mano cuando se elige la fuente o cuando falta un
  glifo. Las caras se versionan; no se generan en cada compilación.

---

## 5. Referencia: el script que hay que sustituir

Esto es lo que produjo las caras de `faces/`. **Funciona y está verificado** — es la
especificación ejecutable de lo que `gofont` debe hacer, no un borrador. Las URLs, las
rutas dentro de cada zip y los flags de `pyftsubset` son datos duros: sacarlos de aquí
ahorra descubrirlos otra vez.

```bash
#!/usr/bin/env bash
# Recorta las caras de Inter y Roboto al rango latino.
# Requisitos (Debian 13):  sudo apt install fonttools
set -euo pipefail
cd "$(dirname "$0")"

# Rango latino de Google Fonts: ASCII + Latin-1 (á é í ó ú ñ ü ¿ ¡),
# puntuación tipográfica (“ ” ‘ ’ – —) y el signo € (U+20AC).
UNICODES="U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,\
U+0304,U+0308,U+0329,U+2000-206F,U+2074,U+20AC,U+2122,U+2191,U+2193,\
U+2212,U+2215,U+FEFF,U+FFFD"

INTER_URL="https://github.com/rsms/inter/releases/download/v4.1/Inter-4.1.zip"
ROBOTO_URL="https://github.com/googlefonts/roboto-3-classic/releases/download/v3.016/Roboto_v3.016.zip"

tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT

curl -sL -o "$tmp/inter.zip"  "$INTER_URL"
curl -sL -o "$tmp/roboto.zip" "$ROBOTO_URL"

# Inter: estáticas en extras/ttf.
# Roboto: unhinted — el hinting sólo afecta al rasterizado en pantalla y suma bytes.
unzip -qo "$tmp/inter.zip"  -d "$tmp/src" "extras/ttf/Inter-Regular.ttf" \
    "extras/ttf/Inter-Bold.ttf" "extras/ttf/Inter-Italic.ttf" \
    "extras/ttf/Inter-BoldItalic.ttf"
unzip -qo "$tmp/roboto.zip" -d "$tmp/src" "unhinted/static/Roboto-Regular.ttf" \
    "unhinted/static/Roboto-Bold.ttf" "unhinted/static/Roboto-Italic.ttf" \
    "unhinted/static/Roboto-BoldItalic.ttf"
find "$tmp/src" -name '*.ttf' -exec mv {} "$tmp/" \;

for style in Regular Bold Italic BoldItalic; do
    for family in Inter Roboto; do
        # Un solo TTF por cara, para web y PDF a la vez. Se conservan kern/liga/clig:
        # el navegador los aplica y el motor de PDF los ignora sin romper nada.
        pyftsubset "$tmp/$family-$style.ttf" --unicodes="$UNICODES" \
            --layout-features=kern,liga,clig --output-file="$family-$style.ttf"
    done
done
```

### Dos correcciones respecto a la versión original

La primera versión de este script generaba **dos** salidas por cara —un TTF sin tablas
de layout para PDF y un WOFF2 para web— y hay que no reintroducirlas:

1. **Un solo TTF, no TTF + WOFF2.** El PDF se genera en el frontend, así que pide el
   mismo archivo que la página ya descargó: acierto de caché, no petición nueva. Medido
   sobre Roboto Regular, servido con brotli: **16.132 B compartido contra 27.890 B** con
   dos formatos, y la mitad de peticiones.
2. **`--layout-features=kern,liga,clig`, no `--layout-features=''`.** Descartar las
   tablas ahorra ~4 KB por cara, pero deja el texto **sin kerning en la web**. Sólo era
   defendible cuando el TTF era exclusivo del PDF, que ignora esas tablas.

> Una advertencia que costó una medición errónea: **recortar un archivo ya recortado no
> avisa**. `pyftsubset` acepta la entrada y produce una salida plausible, así que un
> subset del subset parece correcto y ha perdido tablas en silencio. Partir siempre del
> original descargado. `gofont` debe hacer lo mismo, y es un buen argumento para que
> descargue en vez de aceptar una ruta local.

---

## 6. Verificación

1. `gofont Roboto` en un proyecto limpio deja cuatro `.ttf` con los nombres exactos que
   `Family.Face(Style)` deriva, más el `README.md` de procedencia.
2. Las cuatro caras cubren el español y `€`; la cursiva difiere de la recta.
3. `gofont Robotoo` no descarga nada: nombra el error y lista las familias disponibles.
4. Sin `pyftsubset` instalado, el fallo dice qué instalar. No produce un archivo a medias.
5. `grep -rn "\[\]byte\|os\.\|embed" *.go` en la raíz del módulo sigue vacío.
6. Compila bajo `GOOS=js GOARCH=wasm` — el paquete `font`, no `cmd/`.
7. `gotest`.

---

## 7. Decisiones pendientes

1. ¿Opción A (envolver `pyftsubset`) o C (sólo caras pre-recortadas, sin herramienta)?
2. ¿Qué familias entran en el catálogo cerrado? Hoy `faces/` tiene Roboto e Inter.
3. ¿El binario se llama `gofont`, coherente con `gonew`/`gotest`/`gopush`?
