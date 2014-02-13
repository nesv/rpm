
Summary: The Go Language
Name: golang

%define goversion 1.1beta2
%define goroot %{_libdir}/go%{goversion}
%define alternatives_priority 1098   # 1.1b2 -> 1100 - 2

Version: %goversion
Release: 0%{?dist}
Epoch: 1
# source is just the distributed source package
# recompressed and the root dir renamed
Source: golang-1.1beta2.src.tar.xz

License: BSD
# 3-clause

# Group: 
URL: http://golang.org
BuildRoot: %{_tmppath}/%{name}-%{version}-root
# BuildRequires:
# Requires:

%description
Go is an open source programming environment that makes it easy to build simple, reliable, and efficient software.

%prep

%setup -q -n %{name}-%{version}


mkdir -p $RPM_BUILD_ROOT/%{_bindir}
install -d $RPM_BUILD_ROOT/%{goroot}
install -d $RPM_BUILD_ROOT/%{goroot}

%build
pushd src
GOROOT_FINAL=%{goroot} CGO_ENABLED=1 ./make.bash
popd

cp bin/go $RPM_BUILD_ROOT/%{_bindir}/go%{goversion}
cp bin/gofmt $RPM_BUILD_ROOT/%{_bindir}/gofmt%{goversion}
cp bin/godoc $RPM_BUILD_ROOT/%{_bindir}/godoc%{goversion}

cp -r . $RPM_BUILD_ROOT/%{goroot}


# kick batch files and rc files out, since we're not on Plan9 or Windows
rm -f $RPM_BUILD_ROOT/%{goroot}/src/*.rc
rm -f $RPM_BUILD_ROOT/%{goroot}/src/*.bat

# do we even need bin? nah
rm -rf $RPM_BUILD_ROOT/%{goroot}/bin
#ln -sf /usr/bin $RPM_BUILD_ROOT/%{goroot}/bin

%check
# bootstrap goroot
GOROOT=`pwd`
pushd src
false && PATH=$GOROOT/bin:%{_prefix}/bin CGO_ENABLED=1 ./run.bash
popd

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT

%files
%defattr(0644,root,root,0755)

%attr(0755,-,-) %{_bindir}/go%{goversion}
%attr(0755,-,-) %{_bindir}/gofmt%{goversion}
%attr(0755,-,-) %{_bindir}/godoc%{goversion}

%{goroot}/api
%{goroot}/doc
%{goroot}/include
%{goroot}/lib
%{goroot}/misc
%dir %{goroot}/pkg
%attr(0755,-,-) %{goroot}/pkg/tool
%attr(-,-,-) %{goroot}/pkg/obj
%attr(-,-,-) %{goroot}/pkg/linux_amd64
%{goroot}/src
%{goroot}/test

%doc %{goroot}/AUTHORS
%doc %{goroot}/CONTRIBUTORS
%doc %{goroot}/favicon.ico
%doc %{goroot}/LICENSE
%doc %{goroot}/PATENTS
%doc %{goroot}/README
%doc %{goroot}/robots.txt
%doc %{goroot}/VERSION
%docdir %{goroot}/api
%docdir %{goroot}/doc

%post
alternatives --install %{_libdir}/go gopath %{goroot} %{alternatives_priority}
alternatives --install %{_bindir}/go go %{_bindir}/go%{goversion} %{alternatives_priority}
alternatives --install %{_bindir}/gofmt gofmt %{_bindir}/gofmt%{goversion} %{alternatives_priority}
alternatives --install %{_bindir}/godoc godoc %{_bindir}/godoc%{goversion} %{alternatives_priority}

%preun
alternatives --remove gopath %{goroot}
alternatives --remove go %{_bindir}/go%{goversion}
alternatives --remove gofmt %{_bindir}/gofmt%{goversion}
alternatives --remove godoc %{_bindir}/godoc%{goversion}


%changelog
* Mon Apr 22 2013 Dave Carlson <thecubic@thecubic.net> 1.1beta2-0
- inital version

