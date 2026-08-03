package font

// Declaration is what a project declares in its font.go: the family and the
// subfolder where its faces live. It is built only with Declare, so a partial
// declaration cannot exist.
type Declaration struct {
	family Family
	dir    string
}

// Declare fixes the family and the subfolder where its faces live.
func Declare(family Family, dir string) Declaration {
	return Declaration{family: family, dir: dir}
}

// Family returns the declared typeface name.
func (d Declaration) Family() Family {
	return d.family
}

// Dir returns the declared subfolder, exactly as given.
func (d Declaration) Dir() string {
	return d.dir
}
