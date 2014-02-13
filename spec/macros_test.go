package spec

import "testing"

func TestNewMacro(t *testing.T) {
	var tms = map[RPMMacro]RPMMacro{
		NewMacro("name", "go", false):     RPMMacro{Name: "name", Value: "go", IsGlobal: false},
		NewMacro("_prefix", "/usr", true): RPMMacro{Name: "_prefix", Value: "/usr", IsGlobal: true}}

	for m, n := range tms {
		if m.Name != n.Name || m.Value != n.Value || m.IsGlobal != n.IsGlobal {
			t.Errorf("FAILED\n\t%+v\n\t%+v", m, n)
		}
	}
}

func TestMacroString(t *testing.T) {
	m := NewMacro("name", "go", false)
	testString := "%define name go"
	if m.String() != testString {
		t.Errorf("%q != %q", m.String(), testString)
	}
}

func TestParseMacroDefinitions(t *testing.T) {
	macros := map[string]map[string]bool{
		"booking_repo":                       {"base": false},
		"_use_internal_dependency_generator": {"0": false},
		"__find_requires":                    {"%{nil}": false},
		"debug_package":                      {"%{nil}": true},
		"__spec_install_post":                {"/usr/lib/rpm/check-rpaths /usr/lib/rpm/check-buildroot /usr/lib/rpm/brp-compress": true},
	}

	ems, _ := NewMacroSet()
	for name, vg := range macros {
		for value, global := range vg {
			ems[name] = NewMacro(name, value, global)
		}
	}

	tms := parseMacroDefinitions([]byte(testSpec))

	for name, macro := range ems {
		if _, ok := tms[name]; !ok {
			t.Errorf("%q not found", name)
		} else if !macro.Equals(tms[name]) {
			t.Errorf("Parsed macro %q is different than expected", name)
		}
	}
}

func TestNewMacroSetWithPaths(t *testing.T) {
	return
}
