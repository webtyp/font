package font

import "testing"

// TestConsumerPath walks the real path of a consumer: declare a family, derive
// the four faces, and build the file names tinywasm/pdf registers and
// tinywasm/assetmin serves — as plain string concatenation, the way a consumer
// writes it. If writing this test were awkward, the API would be awkward.
func TestConsumerPath(t *testing.T) {
	d := Declare("Roboto", "fonts/")

	want := []struct {
		style Style
		name  string
	}{
		{Regular, "Roboto"},
		{Bold, "Roboto-Bold"},
		{Italic, "Roboto-Italic"},
		{BoldItalic, "Roboto-BoldItalic"},
	}

	seen := make([]string, 0, len(want))
	for _, w := range want {
		face := d.Family().Face(w.style)
		if face != w.name {
			t.Errorf("Face(%v) = %q, want %q", w.style, face, w.name)
		}

		pdf := d.Dir() + face + ".ttf"
		if pdf != "fonts/"+w.name+".ttf" {
			t.Errorf("pdf line = %q, want %q", pdf, "fonts/"+w.name+".ttf")
		}

		web := d.Dir() + face + ".woff2"
		if web != "fonts/"+w.name+".woff2" {
			t.Errorf("web line = %q, want %q", web, "fonts/"+w.name+".woff2")
		}

		seen = append(seen, face)
	}

	for i, a := range seen {
		for _, b := range seen[i+1:] {
			if a == b {
				t.Errorf("faces are not pairwise distinct: %q repeated", a)
			}
		}
	}
}

// TestFaceDeterministic proves that deriving the same style twice yields the
// same name, and that an out-of-range Style degrades to the regular face.
func TestFaceDeterministic(t *testing.T) {
	f := Family("Roboto")
	for _, s := range []Style{Regular, Bold, Italic, BoldItalic} {
		if a, b := f.Face(s), f.Face(s); a != b {
			t.Errorf("Face(%v) not deterministic: %q vs %q", s, a, b)
		}
	}
	if got := f.Face(Style(9)); got != f.Face(Regular) {
		t.Errorf("Face(9) = %q, want %q", got, f.Face(Regular))
	}
}

// TestDeclarationAccessors proves the declared family and dir are returned
// exactly as given, without added or stripped separators.
func TestDeclarationAccessors(t *testing.T) {
	d := Declare("Roboto", "fonts/")
	if d.Family() != "Roboto" {
		t.Errorf("Family() = %q, want %q", d.Family(), "Roboto")
	}
	if d.Dir() != "fonts/" {
		t.Errorf("Dir() = %q, want %q", d.Dir(), "fonts/")
	}
}
