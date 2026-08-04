---
PLAN: "fix: un solo nombre para la cara regular, y un Dir() con origen definido"
TAG: v0.1.0
---
## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.**

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 4 | One way to do each thing | Una cara tiene **un** nombre de archivo. Hoy hay dos que funcionan. |
| 7 | Self-describing signatures | `Dir()` no dice respecto a qué es relativo, y cada consumidor supone otra cosa. |
| 9 | Lego pieces, never forks | `pdf` está parcheando esta derivación con fallbacks. El arreglo va aquí, no allí. |
| 5 | Minimal surface | La raíz es identidad pura y cruza a WASM entera: lo que no cruce vive en un subpaquete (§4). |

Este módulo es la **única** derivación de nombres de cara del ecosistema. `pdf` la usa
para leer del disco y del `fetch`; `assetmin` la usará para copiar y publicar URLs. Si la
derivación admite dos respuestas, los consumidores divergen sin que nadie lo note.

---

## 1. `Face(Regular)` devuelve un nombre que el ecosistema no usa

```go
func (f Family) Face(s Style) string {
    case Regular: return string(f)      // → "Roboto"
```

Las otras tres caras llevan sufijo (`-Bold`, `-Italic`, `-BoldItalic`); la regular no.
Consecuencias medibles, hoy, en el repositorio:

- **`faces/` se contradice a sí mismo.** `faces/README.md:7` afirma «el nombre de cada
  archivo es exactamente lo que devuelve `Family.Face(Style)`», y el archivo que hay al
  lado se llama **`Roboto-Regular.ttf`**. La regla es falsa para una de cada cuatro caras.
- **`pdf` no confía en la derivación** y acepta los dos nombres
  (`pdf/document.go:38-43`): prueba `Face(Regular)` y, si falla, `string(f)+"-Regular"`.
  Con ambos funcionando, el dev no puede saber cuál es el correcto — hasta que `assetmin`
  copie usando uno solo y la web quede sin fuente mientras el PDF sí la tiene.

### La decisión

**`Face(Regular)` devuelve `string(f) + "-Regular"`.**

Es convención, no verdad; lo que no es negociable es que sea **una**. Se elige `-Regular`
porque:

1. Es coherente con las otras tres caras: las cuatro se nombran igual y ordenan juntas.
2. Es como Google Fonts nombra sus estáticas, así que el dev que descarga una familia
   cualquiera no renombra nada. Con el nombre pelado renombraría uno de cada cuatro.
3. **Los archivos que este repo ya distribuye (`faces/`) y los de `pdf/fpdf/fonts/` ya se
   llaman así.** Elegir el nombre pelado obligaría a renombrarlos todos.

El dev sigue eligiendo lo que le corresponde —la familia y dónde viven sus archivos, que
es exactamente lo que declara— y no elige el esquema de nombres: eso fue deliberado
cuando `LoadDeclared(d)` sustituyó a `LoadTypeface(cuatro rutas)`, para que nadie
escribiera cuatro nombres a mano ni olvidara uno.

---

## 2. `Dir()` no dice respecto a qué es relativo

`Declare("Roboto", "fonts/")` y `Dir()` devuelve `"fonts/"`. ¿Relativo a qué?

El propio README ya se contradice: dice «copia las caras a `config/fonts/`» y en el
ejemplo de arriba declara `"fonts/"`. Y cada consumidor supone algo distinto: `pdf` en
CLI lo trata como ruta de disco relativa al *working directory*
(`pdf/cmd/demo/main.go:16` usa `"fpdf/fonts/"`), y `assetmin` lo necesita relativo a
`RootDir` para copiar.

**Regla, escrita en la firma y en las specs:** `Dir()` es *dónde están las caras, en el
medio de quien lo lee*. Concretamente:

| Consumidor | Medio | Origen del que cuelga |
|---|---|---|
| `assetmin` (build) | disco | la raíz del proyecto (`Config.RootDir`) |
| `pdf` en CLI | disco | el working directory |
| `pdf` en WASM | HTTP | el origen de la página |

Lo que **no** varía nunca es `Family()`: ése es el origen único que garantiza que la web
y el PDF se vean iguales. Documentar esa asimetría explícitamente —`Family()` es global,
`Dir()` es por medio— y dejar de sugerir que hay una ruta canónica.

No cambia el tipo ni la firma: cambia la documentación, que hoy calla lo único que el
lector necesita saber.

---

## 3. Cambios

1. **`font.go`**: `Face(Regular)` → `string(f) + "-Regular"`. El comentario de `Face`
   sigue diciendo «sin extensión» y añade que las cuatro llevan sufijo.
2. **`font_test.go`**: la tabla de derivación pasa a `Roboto-Regular`. Se mantiene la
   invariante de que los cuatro nombres son distintos entre sí y que un `Style` fuera de
   los cuatro constantes cae en el resultado de `Regular`.
3. **`docs/SPECS.md` §2**: la tabla de derivación (`Regular` → `Roboto-Regular`) y la
   regla 3, que hoy dice «`Regular` returns the family untouched».
4. **`docs/SPECS.md` §3**: añadir la regla de `Dir()` de §2 de este plan.
5. **`docs/EMPEZAR.md`**: líneas 45, 105 y 184 muestran `Roboto` / `Roboto.ttf` /
   `fonts/Roboto.ttf`. Actualizar las tres y el árbol de directorios.
6. **`README.md`**: hoy dice «*Copy the ones you want into your project's
   `config/fonts/`*», y eso sugiere que las caras **salen de `faces/`**. No es así: el dev
   las consigue donde quiera —las descarga, las compra, las tiene de antes— y este módulo
   no tiene nada que decir sobre su procedencia.

   Lo que el README debe declarar son las **dos únicas cosas que este módulo exige**, y
   nada más:

   - el directorio es el que el dev pasa a `Declare(...)` — él elige cuál, y `Dir()` es
     relativo a la raíz del proyecto (§2);
   - los cuatro archivos se llaman exactamente lo que devuelve `Face(Style)`, con `.ttf`.

   `faces/` se menciona **sólo** como conveniencia opcional —«si no tienes una familia
   elegida, aquí hay dos verificadas»—, nunca como el origen. Y el ejemplo `Declare("Roboto",
   "fonts/")` debe usar el mismo directorio que el resto de la documentación, en vez de
   contradecirlo.

7. **`faces/README.md`**: su afirmación —«el nombre de cada archivo es exactamente lo que
   devuelve `Face(Style)`»— pasa a ser cierta. Verificar que los ocho archivos coinciden ya
   con la derivación nueva: **no hay que renombrar ninguno**.

**Lo que este plan NO hace:** no toca `pdf`. Sus **tres** ramas de fallback se borran en
`pdf/docs/PLAN.md`, después de publicar esto.

---

## 4. La raíz es identidad; lo que no cruza a WASM vive en un subpaquete

### 4.1 Hoy no hay nada que sacar — y ése es justo el riesgo

Medición, no impresión: la raíz son **62 líneas en dos archivos**, con **cero imports** y
cero build tags. Y todo lo que hay se usa en el navegador: el PDF se genera en el
frontend, así que `pdf.LoadDeclared` corre en WASM y llama a `Family()`, `Dir()` y
`Face(Style)`. No hay una sola línea de backend que mover.

Así que el trabajo no es reorganizar código: es **impedir que entre**. `docs/SPECS.md` ya
declara la invariante —«No `[]byte`, `os.` or `embed` anywhere in the package source or
tests»— y **nada la comprueba**. Es una regla que hay que recordar, y el harness dice
literalmente que si hay que recordarla, es un hueco: *«Things you "have to remember" …
that is a hole in the harness; close it with types or a single path, not with prose.»*

Un módulo de 62 líneas es exactamente donde esto se cuela: alguien añade un `os.ReadFile`
«sólo para validar que la cara existe», nadie lo nota en la revisión, y a partir de ese
commit todos los binarios WASM del ecosistema cargan `os` y su parte de runtime.

### 4.2 La prueba que cierra el hueco

`TestRootIsWasmSafe`, en la raíz. Lee los `.go` **de la raíz** (no los de subpaquetes) y
falla si aparece:

- cualquier `import` — hoy son cero, y el día que uno haga falta, cambiar este test es la
  decisión explícita que lo autoriza;
- `os.`, `embed`, `[]byte`, `map[`;
- una directiva `//go:build`, en cualquiera de sus dos formas.

Precedente exacto y espejo: `css` tiene `TestPackageIsBuildTimeOnly` (`tokens_test.go`)
para la invariante contraria — ese paquete **nunca** debe cruzar. Éste **siempre** cruza,
así que se le exige lo simétrico.

### 4.3 Dónde va lo que no cruza

| Qué | Dónde | Por qué |
|---|---|---|
| El CLI de subsetting | `cmd/gofont/` | `cmd/` no lo importa ningún consumidor; ya planificado en `docs/GOFONT_PLAN.md` |
| Cualquier E/S, validación o conversión futura | un subpaquete con nombre propio | Go compila por paquete: lo que el front no importa no entra en el binario |
| Las caras `.ttf` | `faces/` | son archivos, no código; ningún `.go` los embebe |

**Regla para consumidores, que va en `README.md` y en `docs/SPECS.md`:** importar la raíz
es gratis y seguro desde cualquier medio; **importar un subpaquete de `font` es una
decisión de backend**, y el archivo que lo haga lleva `//go:build !wasm`. La raíz nunca
importa a sus propios subpaquetes — esa flecha sólo va hacia afuera.

Nota histórica: el primer candidato a subpaquete fue `font/assets` (el `FontProcessor` que
copiaría las caras). Se descartó por otra razón —la entrega es de `assetmin`, que ya es
quien sirve— pero de haber existido, éste es el sitio donde habría vivido, y no la raíz.

---

## 5. Verificación

1. `Family("Roboto").Face(Regular) == "Roboto-Regular"`, y las cuatro caras coinciden
   **una a una** con los archivos reales de `faces/` — un test que recorre los cuatro
   `Style` y hace `os.Stat` no puede escribirse aquí (§ invariantes: nada de `os.`), así
   que la comprobación va en la revisión: cuatro nombres, cuatro archivos, mismo string.
2. `TestRootIsWasmSafe` en verde, y **falla** si se le añade a la raíz un archivo con un
   `import "os"` — comprobarlo a mano una vez antes de dar el test por bueno. Un test de
   invariante que no se ha visto fallar no prueba nada.
3. La raíz sigue con **cero imports**, cero build tags y ningún `map[`.
4. `gotest`.
4. Ninguna mención a `Roboto.ttf` (nombre pelado) queda en el repositorio:
   `grep -rn "Roboto\.ttf\|Face(font.Regular)" .` sólo devuelve el nombre nuevo.

Es una ruptura de API deliberada y por eso sube a **v0.1.0**: cualquier proyecto que
tuviera su cara regular con el nombre pelado tiene que renombrar un archivo, y lo
descubrirá con un error que nombra el archivo que falta — no con un PDF sin cursivas.
