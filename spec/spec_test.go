package spec

import (
	"bytes"
	"fmt"
	"testing"
)

var (
	testSpec = `
%define booking_repo base
# Specfile for go. See README for instructions

Name:          go
Version:       1.1
Release:       1%{?dist}
Summary:       Go compiler and tools
Group:         Development/Languages
License:       BSD
URL:           http://golang.org/
Source0:       http://go.googlecode.com/files/%{name}%{version}.src.tar.gz
Patch0:        diff0.patch
BuildRoot:     %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
BuildRequires: ed
BuildRequires: bison
BuildRequires: mercurial
%define _use_internal_dependency_generator 0
%define __find_requires %{nil}
%global debug_package %{nil}
%global __spec_install_post /usr/lib/rpm/check-rpaths /usr/lib/rpm/check-buildroot /usr/lib/rpm/brp-compress

%ifarch %ix86
    %global GOARCH 386
%endif
%ifarch    x86_64
    %global GOARCH amd64
%endif

%description
Go is a systems programming language that aims to be both fast and convenient.

%package  vim
Summary:  go syntax files for vim
Group:    Applications/Editors
Requires: vim-common
Requires: %{name} = %{version}-%{release}

%description vim
Go syntax for vim.

%package  emacs
Summary:  go syntax files for emacs
Group:    Applications/Editors
Requires: emacs-common
Requires: %{name} = %{version}-%{release}

%description  emacs
Go syntax for emacs.

%prep
%setup -q -n go

%build
GOSRC="$(pwd)"
GOROOT="$(pwd)"
GOROOT_FINAL=%{_libdir}/go
GOOS=linux
GOBIN="$GOROOT/bin"
GOARCH="%{GOARCH}"
export GOARCH GOROOT GOOS GOBIN GOROOT_FINAL
export MAKE=%{__make}

mkdir -p "$GOBIN"
cd src

LC_ALL=C PATH="$PATH:$GOBIN" ./all.bash

%install
rm -rf %{buildroot}

GOROOT_FINAL=%{_libdir}/go
GOROOT="%{buildroot}%{_libdir}/go"
GOOS=linux
GOBIN="$GOROOT/bin"
GOARCH="%{GOARCH}"
export GOARCH GOROOT GOOS GOBIN GOROOT_FINAL

install -Dm644 misc/bash/go %{buildroot}%{_sysconfdir}/bash_completion.d/go
install -Dm644 misc/emacs/go-mode-load.el %{buildroot}%{_datadir}/emacs/site-lisp/go-mode-load.el
install -Dm644 misc/emacs/go-mode.el %{buildroot}%{_datadir}/emacs/site-lisp/go-mode.el
install -Dm644 misc/vim/syntax/go.vim %{buildroot}%{_datadir}/vim/vimfiles/syntax/go.vim
install -Dm644 misc/vim/ftdetect/gofiletype.vim %{buildroot}%{_datadir}/vim/vimfiles/ftdetect/gofiletype.vim
install -Dm644 misc/vim/ftplugin/go/fmt.vim %{buildroot}%{_datadir}/vim/vimfiles/ftplugin/go/fmt.vim
install -Dm644 misc/vim/ftplugin/go/import.vim %{buildroot}%{_datadir}/vim/vimfiles/ftplugin/go/import.vim
install -Dm644 misc/vim/indent/go.vim %{buildroot}%{_datadir}/vim/vimfiles/indent/go.vim

mkdir -p $GOROOT/{misc,lib,src}
mkdir -p %{buildroot}%{_bindir}/

cp -ar pkg include lib bin $GOROOT
cp -ar src/pkg src/cmd $GOROOT/src
cp -ar misc/cgo $GOROOT/misc

ln -sf %{_libdir}/go/bin/go %{buildroot}%{_bindir}/go
ln -sf %{_libdir}/go/bin/godoc %{buildroot}%{_bindir}/godoc
ln -sf %{_libdir}/go/bin/gofmt %{buildroot}%{_bindir}/gofmt

ln -sf %{_libdir}/go/pkg/tool/linux_%{GOARCH}/cgo %{buildroot}%{_bindir}/cgo
ln -sf %{_libdir}/go/pkg/tool/linux_%{GOARCH}/ebnflint %{buildroot}%{_bindir}/ebnflint

%ifarch %ix86
for tool in 8a 8c 8g 8l; do
%else
for tool in 6a 6c 6g 6l; do
%endif
ln -sf %{_libdir}/go/pkg/tool/linux_%{GOARCH}/$tool %{buildroot}%{_bindir}/$tool
done

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%doc AUTHORS CONTRIBUTORS LICENSE README doc/*
%{_libdir}/go
%ifarch %ix86
%{_bindir}/8*
%else
%{_bindir}/6*
%endif
%{_bindir}/cgo
%{_bindir}/ebnflint
%{_bindir}/go*
%{_sysconfdir}/bash_completion.d/go

%files vim
%defattr(-,root,root,-)
%{_datadir}/vim/vimfiles/ftdetect/gofiletype.vim
%{_datadir}/vim/vimfiles/ftplugin/go/fmt.vim
%{_datadir}/vim/vimfiles/ftplugin/go/import.vim
%{_datadir}/vim/vimfiles/indent/go.vim
%{_datadir}/vim/vimfiles/syntax/go.vim

%files emacs
%defattr(-,root,root,-)
%{_datadir}/emacs/site-lisp/go-mode*.el

%changelog
* Fri Dec 07 2012 Taylor Goodwill <tgoodwill@synacor.com> - 1.0.3-1_synacor
- Initial Synacor Build
`
	parsedSpec *SpecFile
)

func init() {
	var err error
	parsedSpec, err = ParseString(testSpec)
	if err != nil {
		panic("parsing of test spec failed")
	}
}

func TestParseName(t *testing.T) {
	ename := "go"
	pname := parsedSpec.Name()

	t.Logf("expecting %q", ename)

	if pname != ename {
		t.Errorf("failed to parse package name from spec; got '%s' expected '%s'", pname, ename)
	}
	return
}

func TestParseVersion(t *testing.T) {
	version := "1.1"
	pver := parsedSpec.Version()

	t.Logf("expecting %q", version)

	if pver != version {
		t.Errorf("wrong version; got %q wanted %q", pver, version)
	}
	return
}

func TestSubpackages(t *testing.T) {
	esubpackages := []string{"vim", "emacs"}
	psubpackages := parsedSpec.Subpackages()

	t.Logf("expecting %q", esubpackages)

	if len(psubpackages) == 0 {
		t.Error("failed to parse subpackage names; got nothing")
	}

	if fmt.Sprintf("%q", psubpackages) != fmt.Sprintf("%q", esubpackages) {
		t.Errorf("wrong subpackage matches; got %q wanted %q", psubpackages, esubpackages)
	}

	return
}

func TestSummary(t *testing.T) {
	esummary := "Go compiler and tools"
	t.Logf("expecting '%q'", esummary)

	psummary := parsedSpec.Summary()
	if psummary != esummary {
		t.Errorf("failed to parse summary; got '%q' wanted '%q'", psummary, esummary)
	}

	return
}

func TestSources(t *testing.T) {
	esources := map[string]string{"0": "http://go.googlecode.com/files/go1.1.src.tar.gz"}
	psources := parsedSpec.Sources()

	t.Logf("expecting %q", esources)

	if len(psources) == 0 {
		t.Error("failed to parse sources; got nothing")
	}

	if fmt.Sprintf("%q", psources) != fmt.Sprintf("%q", esources) {
		t.Errorf("wrong source matches; got %q wanted %q", psources, esources)
	}

	return
}

func TestPatches(t *testing.T) {
	epatches := map[string]string{"0": "diff0.patch"}
	t.Logf("expecting %q", epatches)

	ppatches := parsedSpec.Patches()
	if fmt.Sprintf("%q", ppatches) != fmt.Sprintf("%q", epatches) {
		t.Errorf("wrong patch matches; got %q wanted %q", ppatches, epatches)
	}

	return
}

func TestBuildRequires(t *testing.T) {
	ebreqs := []string{"ed", "bison", "mercurial"}
	t.Logf("expecting %q", ebreqs)

	pbreqs := parsedSpec.BuildRequires()
	if fmt.Sprintf("%q", pbreqs) != fmt.Sprintf("%q", ebreqs) {
		t.Errorf("wrong buildrequires matches; got %q wanted %q", pbreqs, ebreqs)
	}

	return
}

func TestRequires(t *testing.T) {
	ereqs := []string{"vim-common", "go = 1.1-1", "emacs-common"}
	t.Logf("expecting %q", ereqs)

	preqs := parsedSpec.Requires()
	if fmt.Sprintf("%q", preqs) != fmt.Sprintf("%q", ereqs) {
		t.Errorf("wrong requires matches; got %q wanted %q", preqs, ereqs)
	}

	return
}

func TestMacroSubstitutions(t *testing.T) {
	eba := []byte("http://go.googlecode.com/files/go1.1.src.tar.gz")
	sba := []byte(parsedSpec.Sources()["0"])
	subs := map[string]RPMMacro{
		"name":    NewMacro("name", string(parsedSpec.Name()), false),
		"version": NewMacro("version", string(parsedSpec.Version()), false)}

	t.Logf("expecting %q", eba)

	pba, err := parsedSpec.macroSub(sba, subs)

	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(pba, eba) {
		t.Errorf("failed; got %q wanted %q", pba, eba)
	}

	return
}

func TestRelease(t *testing.T) {
	erelease := "1"
	prelease := parsedSpec.Release()

	t.Logf("expecting %q", erelease)

	if len(prelease) == 0 {
		t.Error("failed to parse release; got nothing")
	}

	if prelease != erelease {
		t.Errorf("wrong release version; got %q wanted %q", prelease, erelease)
	}

	return
}
