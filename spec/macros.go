package spec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

var (
	reDefineMacro = regexp.MustCompile(`%(define|global)\s+([^\s]+)\s+([^\r\n]+)`)
)

/*
An RPMMacro represents a user-defined, or user-accessible, macro that would be
defined in a spec file.
*/
type RPMMacro struct {
	Name     string
	Value    string
	IsGlobal bool
}

/*
Creates a new `RPMMacro`.
*/
func NewMacro(name, value string, global bool) RPMMacro {
	return RPMMacro{Name: name, Value: value, IsGlobal: global}
}

/*
Returns the macro formatted so it can be written out to a file and loaded back
in.
*/
func (m RPMMacro) String() string {
	var fmtstr string

	if m.IsGlobal {
		fmtstr = "%%global %s %s"
	} else {
		fmtstr = "%%define %s %s"
	}

	return fmt.Sprintf(fmtstr, m.Name, m.Value)
}

/*
Quickly test if the associated `RPMMacro` is equal to the provided one, in
values only.
*/
func (m RPMMacro) Equals(n RPMMacro) bool {
	if m.Name == n.Name && m.Value == n.Value && m.IsGlobal == n.IsGlobal {
		return true
	}

	return false
}

/*
A MacroSet is a convenience type for holding a list of RPM macros, that are
accessible by the macros' names.
*/
type MacroSet map[string]RPMMacro

/*
A convenience function that lets you merge in the provided *MacroSet. Any
existing macros are overwritten.
*/
func (ms MacroSet) Update(m MacroSet) {
	for name, macro := range m {
		ms[name] = macro
	}
}

/*
Creates a new MacroSet.

Optionally, you can provide several pathnames, by which to load in pre-defined
macros (in example "$HOME/.rpmmacros"). Providing a directory (such as
"/etc/rpm/macros") will result in an error.
*/
func NewMacroSet(paths ...string) (ms MacroSet, err error) {
	ms = make(MacroSet)
	if len(paths) > 0 {
		// If the caller has provided some files they would like to load
		// some macro definitions from, then iterate over the list of
		// provided paths, load each file, and merge it into the main
		// MacroSet.
		var mm MacroSet
		for _, i := range paths {
			mm, err = loadMacroFile(i)
			if err != nil {
				return
			}
			ms.Update(mm)
		}
	}

	return
}

/*
Loads the file specified by `path`, and returns a pointer to a `MacroSet`, and
an error (if something went wrong).
*/
func loadMacroFile(path string) (ms MacroSet, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return
	}

	defer f.Close()

	var b []byte
	b, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}

	ms = parseMacroDefinitions(b)
	return
}

/*
This function takes a byte slice, and parses it, looking for "%define" or
"%global" statements, and returns any discovered macros in a `MacroSet`.
*/
func parseMacroDefinitions(b []byte) (ms MacroSet) {
	ms = make(MacroSet)

	matches := reDefineMacro.FindAllStringSubmatch(string(b), -1)
	if matches == nil {
		// No matches does not necessarily mean that parsing failed.
		return
	}

	for _, match := range matches {
		if len(match) < 4 {
			// Skip any malformed macro definitions.
			continue
		}

		var macro RPMMacro
		if match[1] == "global" {
			macro = NewMacro(match[2], match[3], true)
		} else {
			macro = NewMacro(match[2], match[3], false)
		}

		ms[match[2]] = macro
	}

	return
}

/*
Expands the macros in the provided byte slice, using the macros present in the
provided `*MacroSet`. On a successful invocation, this function will return
a byte slice with all of the macros expanded, and a `nil` error.
*/
func ExpandMacros(b []byte, macros MacroSet) (expanded []byte, err error) {
	err = errors.New("Not implemented")
	return
}
