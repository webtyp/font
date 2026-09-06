# faces/ — tipografías listas para usar

Caras verificadas que un proyecto **copia** a su `config/fonts/`. No se importan desde
Go: son archivos, no código. Ningún `.go` de este módulo las referencia ni las embebe,
así que **no entran en ningún binario**.

El nombre de cada archivo es exactamente lo que devuelve `Family.Face(Style)` más
`.ttf`. Esa correspondencia es la razón de que la carpeta se llame `faces/`:

```go
font.Family("Roboto").Face(font.Bold)   // "Roboto-Bold"  →  faces/Roboto-Bold.ttf
```

---

## Qué hay

| Familia | Caras | Crudo | Servido (brotli) | Licencia |
|---|---|---|---|---|
| **Roboto** | 4 | 115.096 B | **69.978 B** | OFL — `Roboto-LICENSE-OFL.txt` |
| **Inter** | 4 | 209.288 B | 95.725 B | OFL — `Inter-LICENSE-OFL.txt` |

Ambas cumplen lo que el ecosistema exige de una tipografía:

- **TrueType estática.** El motor de `webtyp/pdf` lee `glyf`/`loca`: ignora los ejes
  de una fuente variable y no puede cargar contornos CFF (`.otf`).
- **Cuatro caras reales.** Regular, Bold, Italic y BoldItalic son archivos distintos —
  la cursiva es una cursiva, no la recta inclinada por el motor.
- **Cobertura del español, con `€`.** `áéíóúñü ÁÉÍÓÚÑÜ ¿¡`, comillas tipográficas
  `“” ‘’`, guiones `– —` y el signo `€`.

Verificado generando un PDF con las cuatro caras y extrayendo su texto.

### Cuál elegir

**Roboto.** Pesa un 27% menos servida (69.978 B contra 95.725 B las cuatro caras) y es
lo que Android ya renderiza de forma nativa, así que esos usuarios no perciben cambio y
iOS converge hacia ellos.

Inter está muy bien diseñada para interfaz, pero sus tablas OpenType son bastante más
grandes: 51 KB por cara frente a 28 KB. La diferencia sólo desaparece si se descartan
esas tablas, y no se pueden descartar porque el navegador las usa para el kerning.

---

## Procedencia y recorte

Son **derivados**: la fuente original completa pesa ~400 KB por cara porque trae más de
3.000 glifos (griego, cirílico, vietnamita). Aquí están recortadas al alfabeto latino,
lo que deja Roboto en ~28 KB por cara e Inter en ~51 KB.

| | Origen | Caras usadas |
|---|---|---|
| Roboto | v3.016 — `github.com/googlefonts/roboto-3-classic` | `unhinted/static/` |
| Inter | v4.1 — `github.com/rsms/inter` | `extras/ttf/` |

De Roboto se toman las **unhinted**: el *hinting* sólo afecta al rasterizado en
pantalla y suma bytes que ni el navegador moderno ni el PDF aprovechan.

Rango Unicode conservado — el subset latino de Google Fonts más `€` (U+20AC):

```
U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,U+0304,
U+0308,U+0329,U+2000-206F,U+2074,U+20AC,U+2122,U+2191,U+2193,U+2212,
U+2215,U+FEFF,U+FFFD
```

Comando que las produce, por cara:

```bash
pyftsubset <cara>.ttf --unicodes="<rango de arriba>" \
    --layout-features=kern,liga,clig --output-file=<cara>.ttf
```

Se conservan `kern`, `liga` y `clig` porque el navegador los aplica; el motor de PDF
los ignora sin que eso rompa nada.

> **Si falta un glifo**, esos tres datos —versión, caras y rango— son todo lo que hace
> falta para rehacer el subset. Sin ellos estos archivos serían binarios huérfanos.
---

## Por qué no está DroidSans

Estuvo en `webtyp/pdf` y **no se traslada**, aunque pese menos (83.508 B): tiene sólo
dos caras, le falta el glifo `€` y no trae cursiva real —los consumidores apuntaban los
estilos `"I"`/`"BI"` a los archivos rectos, así que nunca hubo cursiva—. Está
descontinuada desde Android 4.0 (2011).

Un catálogo curado no puede incluir la tipografía que el resto de la documentación usa
como contraejemplo.
