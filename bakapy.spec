Summary: Bakapy backup framework
Name: bakapy
Version: %{version}
Release: %{release}%{dist}
Source0: http://bakapy.org/download/bakapy-%{version}.tar.gz
License: GPLv3
Group: Backup
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-buildroot
Vendor: Subuk

%description
Bakapy backup framework

%prep
%setup -q

%build
./build.sh

%install
install -m 755 -d %{buildroot}/usr/bin
install -m 755 -d %{buildroot}/etc
install -m 755 -d %{buildroot}/etc/init
install -m 755 -d %{buildroot}/etc/bakapy

cp -r bin/ %{buildroot}/usr/
cp -r commands/ %{buildroot}/etc/bakapy/
cp -r bakapy.conf.ex.yaml %{buildroot}/etc/bakapy/bakapy.conf
cp -r jobs.conf.ex.yaml %{buildroot}/etc/bakapy/jobs.conf
cp -f debian/bakapy.upstart %{buildroot}/etc/init/bakapy.conf


%clean
rm -rf $RPM_BUILD_ROOT

%files
%attr(755,root,root) %{_prefix}/bin/bakapy-scheduler
%attr(755,root,root) %{_prefix}/bin/bakapy-run-job
%attr(755,root,root) %{_prefix}/bin/bakapy-show-meta
%config(noreplace) /etc/bakapy
%attr(644,root,root) /etc/init/bakapy.conf
