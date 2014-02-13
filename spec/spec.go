package spec

import (
	"container/set"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	reName          = regexp.MustCompile(`[Nn]ame:\s+([^\n]+)`)
	reVersion       = regexp.MustCompile(`[Vv]ersion:\s+([^\n]+)`)
	reRelease       = regexp.MustCompile(`Release:\s+([^\n]+)`)
	reSummary       = regexp.MustCompile(`Summary:\s+([^\n]+)`)
	reRequires      = regexp.MustCompile(`[^d]Requires:\s+([^\n]+)`)
	reBuildRequires = regexp.MustCompile(`BuildRequires:\s+([^\n]+)`)
	reSubpackage    = regexp.MustCompile(`%package\s+([^\n]+)`)
	reSource        = regexp.MustCompile(`Source(\d{1,3}):\s+([^\n]+)`)
	rePatch         = regexp.MustCompile(`Patch(\d{1,3}):\s+([^\n]+)`)
)

var (
	ErrTooFewSubs  = errors.New("Too few submatches.")
	ErrTooManySubs = errors.New("Too many submatches.")
	ErrNoSubs      = errors.New("No submatches found.")
)

type SpecFile struct {
	raw    []byte
	macros MacroSet
}

func (s *SpecFile) findSubmatch(re *regexp.Regexp, nmatches int) ([]byte, error) {
	matches := re.FindSubmatch(s.raw)
	if matches != nil {
		if len(matches) > nmatches {
			return nil, ErrTooManySubs
		} else if len(matches) < nmatches {
			return nil, ErrTooFewSubs
		}
	} else {
		return nil, ErrNoSubs
	}
	return matches[nmatches-1], nil
}

func (s *SpecFile) findStringSubmatches(re *regexp.Regexp, nmatches int) ([]string, error) {
	var err error
	var matches = make([]string, 0)

	mm := re.FindAllStringSubmatch(string(s.raw), -1)
	if mm != nil {
		for _, match := range mm {
			if len(match) == nmatches {
				matches = append(matches, match[nmatches-1])
			} else if len(match) > nmatches {
				err = errors.New("too many submatches")
			} else if len(match) < nmatches {
				err = errors.New("too few submatches")
			}
		}
	} else {
		err = errors.New("no matches found")
	}

	return matches, err
}

/*
This function performs substitutions on byte arrays where a macro is placed,
in the spec file.

For example, if you see "%{name}" in a string somewhere, this function will
return you a new byte array with all occurrences of "%{name}" replaced with the
value returned by calling `*SpecFile.Name()`.

The first argument to this function is the byte array you would like to apply
the subtitutions to. The second argument, `subs`, is a string-string mapping
where the keys are the macro to search for, and the value is what to replace
the macro with.
*/
func (s *SpecFile) macroSub(src []byte, macros map[string]RPMMacro) ([]byte, error) {
	dest := src
	for _, macro := range macros {
		re, err := regexp.Compile(fmt.Sprintf("%%\\{%s\\}", macro.Name))
		if err != nil {
			return nil, err
		}

		dest = re.ReplaceAll(dest, []byte(macro.Value))
	}
	return dest, nil
}

func (s *SpecFile) BuildRequires() []string {
	breqs, err := s.findStringSubmatches(reBuildRequires, 2)
	if err != nil {
		return nil
	}

	buildreqs := set.New()
	for _, i := range breqs {
		b, err := s.macroSub([]byte(i), s.macros)
		if err != nil {
			panic(err)
		}

		for _, i := range strings.Split(string(b), ",") {
			_ = buildreqs.Add(strings.Trim(i, " "))
		}
	}

	return buildreqs.Members()
}

func (s *SpecFile) Requires() []string {
	reqs, err := s.findStringSubmatches(reRequires, 2)
	if err != nil {
		return nil
	}

	requires := set.New()
	for _, r := range reqs {
		b, err := s.macroSub([]byte(r), s.macros)
		if err != nil {
			panic(err)
		}

		// The next little bit of code, splits up the parsed "Require:" line,
		// on commas, and then trims the whitespace off the ends of the
		// resulting substrings.
		//
		// In the even there are no comma-separated requirements, this clause
		// won't "damage" anything.
		for _, i := range strings.Split(string(b), ",") {
			_ = requires.Add(strings.Trim(i, " "))
		}
	}

	return requires.Members()
}

func (s *SpecFile) Patches() map[string]string {
	patches := make(map[string]string, 0)
	matches := rePatch.FindAllStringSubmatch(string(s.raw), -1)
	if matches != nil {
		for _, match := range matches {
			if len(match) == 3 {
				p, err := s.macroSub([]byte(match[2]), s.macros)
				if err != nil {
					panic(err)
				}

				patches[match[1]] = string(p)
			}
		}
		return patches
	}
	return nil
}

func (s *SpecFile) Sources() map[string]string {
	matches := reSource.FindAllStringSubmatch(string(s.raw), -1)
	if matches != nil {
		sources := make(map[string]string, 0)
		for _, match := range matches {
			if len(match) == 3 {
				src, err := s.macroSub([]byte(match[2]), s.macros)
				if err != nil {
					panic(err)
				}

				sources[match[1]] = string(src)
			}
		}
		return sources
	}
	return nil
}

func (s *SpecFile) Subpackages() []string {
	matches := reSubpackage.FindAllStringSubmatch(string(s.raw), -1)
	if matches != nil {
		subpackages := make([]string, 0)
		for _, match := range matches {
			if len(match) == 2 {
				subpackages = append(subpackages, match[1])
			}
		}
		return subpackages
	}
	return nil
}

/*
Returns the package summary from the spec file.

A zero-length string indicates the spec file failed to be parsed, and may
indicated a malformed spec.
*/
func (s *SpecFile) Summary() string {
	if match, err := s.findSubmatch(reSummary, 2); err != nil {
		return ""
	} else {
		return string(match)
	}
}

/*
Returns the release of the package in the spec file. The "release" can also
be referred to as the "build number", and sometimes has an additional "dist"
tag attached to it. If there is a "dist" macro in the release string,
it will be stripped.

A zero-length string indicates the spec file failed to be parsed, and may
indicated a malformed spec.
*/
func (s *SpecFile) Release() string {
	var rel []byte
	var err error
	var match []byte

	if match, err = s.findSubmatch(reRelease, 2); err != nil {
		return ""
	} else {
		subs := make(map[string]RPMMacro)
		subs["\\??dist"] = NewMacro("\\??dist", "", false)

		rel, err = s.macroSub(match, subs)
		if err != nil {
			return ""
		}
	}

	return string(rel)
}

/*
Returns the version of the package in the spec file. A zero-length string
indicates that the spec file failed to be parsed, and may indicate a malformed
spec.
*/
func (s *SpecFile) Version() string {
	if match, err := s.findSubmatch(reVersion, 2); err != nil {
		return ""
	} else {
		return string(match)
	}
}

/*
Returns the name of the package in the spec file. A zero-length string
indicates that the spec file failed to be parsed, and may indicate a malformed
spec.
*/
func (s *SpecFile) Name() string {
	if match, err := s.findSubmatch(reName, 2); err != nil {
		return ""
	} else {
		return string(match)
	}
}

/*
Raw returns the byte slice containing the spec file data that was provided to
one of the Parse* functions.
*/
func (s *SpecFile) Raw() []byte {
	return s.raw
}

func Parse(data []byte) (*SpecFile, error) {
	var spec = SpecFile{raw: data}

	// Find any declared macros, from within the spec file.
	mdefs := parseMacroDefinitions(data)

	// Check a few of the required macros, and if they have not been
	// defined by a "%define" statement, then parse them out from elsewhere
	// in the spec file.
	if _, ok := mdefs["name"]; !ok {
		mdefs["name"] = NewMacro("name", spec.Name(), false)
	}

	if _, ok := mdefs["version"]; !ok {
		mdefs["version"] = NewMacro("version", spec.Version(), false)
	}

	if _, ok := mdefs["release"]; !ok {
		mdefs["release"] = NewMacro("release", spec.Release(), false)
	}

	spec.macros = mdefs
	return &spec, nil
}

func ParseString(data string) (*SpecFile, error) {
	return Parse([]byte(data))
}

func ParseReader(r io.Reader) (*SpecFile, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return nil, errors.New("error: zero-length data, nothing to parse")
	}

	return Parse(data)
}
