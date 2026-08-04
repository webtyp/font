// Package font names the typeface of a product and derives the names of its
// faces. It carries no bytes, no paths and no medium knowledge: web and PDF
// derive their font names from this single origin.
//
// It travels inside the WASM binary, so it never reads a file.
package font

// Family is the name of the product's typeface. It is the only thing of this
// package that crosses to WASM.
type Family string

// Style names a face. The set is closed: these four, and exactly the four
// tinywasm/pdf registers.
type Style uint8

const (
	Regular Style = iota
	Bold
	Italic
	BoldItalic
)

// Face is a face's file name, derived — never written by hand. Without
// extension: each medium adds its own (.ttf for PDF, .woff2 for web). All
// four styles carry a "-Style" suffix, Regular included.
func (f Family) Face(s Style) string {
	switch s {
	case Regular:
		return string(f) + "-Regular"
	case Bold:
		return string(f) + "-Bold"
	case Italic:
		return string(f) + "-Italic"
	case BoldItalic:
		return string(f) + "-BoldItalic"
	}
	return string(f) + "-Regular"
}
